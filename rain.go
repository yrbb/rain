package rain

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"

	"github.com/yrbb/rain/pkg/database"
	"github.com/yrbb/rain/pkg/logger"
	"github.com/yrbb/rain/pkg/redis"
)

func init() {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}
}

type Rain struct {
	pid        int
	started    int32
	closed     int32
	exitCh     chan os.Signal
	isShowHelp bool
	isServer   bool

	config *Config

	cmd      *cobra.Command
	database *database.Database
	redis    *redis.Redis
	worker   *ants.Pool
	engine   *gin.Engine
	server   *http.Server
	watcher  *fsnotify.Watcher

	beforeStart    []func()
	beforeStop     []func()
	onConfigUpdate func(config *Config)
}

func New() (*Rain, error) {
	p := &Rain{
		cmd:    RootCmd,
		pid:    os.Getpid(),
		exitCh: make(chan os.Signal, 10),
	}

	registerServerCommand(p)
	registerVersionCommand(p)

	cmd, args, err := p.cmd.Find(os.Args[1:])
	if err != nil {
		return nil, err
	}

	p.checkIsHelpCommand(cmd, args)

	if !p.isShowHelp {
		cmd.ParseFlags(args)

		file, err := cmd.Flags().GetString("config")
		if err != nil || strings.Contains(file, "=") {
			return nil, fmt.Errorf("配置文件参数异常")
		}

		if file == "" {
			file = "config.toml"
		}

		if p.config, err = parseConfig(file); err != nil {
			return nil, err
		}

		p.initLogger()
		p.initConfigWatcher(file, p.configUpdate)
	}

	gin.SetMode(gin.ReleaseMode)
	p.engine = gin.New()

	rainIns = p

	return p, nil
}

func (p *Rain) initLogger() {
	logger.Init(
		p.config.Project,
		p.config.Logger.Level,
		p.config.Logger.Path,
		p.config.Logger.SplitTime,
	)
	logger.SetDebug(p.config.Debug)
}

func (p *Rain) initConfigWatcher(file string, callback func(*Config)) {
	var err error
	p.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.M().Error("启动配置监听器异常")
		os.Exit(1)
	}

	watchConfig(p.watcher, file, callback)
}

func (p *Rain) beforeStartCallback() {
	showRainInfo()

	if p.isShowHelp {
		return
	}

	if err := p.initComponents(); err != nil {
		logger.M().Error(err.Error())
		os.Exit(1)
	}

	if l := len(p.beforeStart); l > 0 {
		for i := 0; i < l; i++ {
			p.beforeStart[i]()
		}
	}
}

func (p *Rain) serverReadyCheck() {
	if !p.isServer {
		return
	}

	time.Sleep(time.Millisecond * 200)

	for i := 0; i <= 10; i++ {
		if i == 10 {
			p.stop()
			logger.M().Error("HTTP 服务启动异常")
			os.Exit(1)
		}

		time.Sleep(time.Millisecond * 10)

		if serverIsReady(p.config.Server.Listen) {
			break
		}
	}

	atomic.StoreInt32(&p.started, 1)
}

func (p *Rain) Run() {
	wrapCommand(p)

	if err := p.cmd.Execute(); err != nil {
		logger.M().Error(err.Error())
		p.stop()
	}
}

func (p *Rain) configUpdate(cfg *Config) {
	logger.SetDebug(cfg.Debug)

	if p.redis != nil {
		p.redis.SetDebug(cfg.Debug)
		p.redis.UpdateConfig(cfg.Redis)
	}

	if p.database != nil {
		p.database.SetDebug(cfg.Debug)
		p.database.UpdateConfig(cfg.Database)
	}

	if p.onConfigUpdate != nil {
		p.onConfigUpdate(cfg)
	}
}

func (p *Rain) listenSignals() {
	signal.Notify(p.exitCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-p.exitCh

		if sig == syscall.SIGINT {
			fmt.Println("")
		}

		logger.M().Info(fmt.Sprintf("收到信号: %s, Pid: %d", sig.String(), p.pid))

		if p.server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), p.config.Server.StopTimeout)
			if err := p.server.Shutdown(ctx); err != nil {
				logger.M().Error("HTTP 服务停止异常", slog.String("error", err.Error()))
			} else {
				logger.M().Info("HTTP 服务停止")
			}
			cancel()
		}

		p.stop()
	}
}

func (p *Rain) stop() {
	if p.isShowHelp {
		os.Exit(0)
	}

	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return
	}

	if l := len(p.beforeStop); l > 0 {
		for i := l - 1; i >= 0; i-- {
			p.beforeStop[i]()
		}
	}

	p.cleanComponents()

	logger.M().Info(fmt.Sprintf("服务退出, Pid: %d", p.pid))
	logger.Close()

	os.Exit(0)
}

func (p *Rain) checkIsHelpCommand(cmd *cobra.Command, args []string) {
	if name := cmd.Name(); name == "help" || name == "version" || name == "rain" {
		p.isShowHelp = true
		return
	}

	for _, v := range args {
		if v == "-h" || v == "help" {
			p.isShowHelp = true
			return
		}
	}
}

func (p *Rain) initComponents() error {
	err := p.initWorker()
	if err != nil {
		return err
	}

	p.redis, err = redis.New(p.config.Redis)
	if err != nil {
		return err
	}
	p.redis.SetDebug(p.config.Debug)

	p.database, err = database.New(p.config.Database)
	if err != nil {
		return err
	}
	p.database.SetDebug(p.config.Debug)

	return nil
}

func (p *Rain) cleanComponents() {
	if p.database != nil {
		p.database.Close()
		logger.M().Info("清理数据库资源")
	}

	if p.redis != nil {
		p.redis.Close()
		logger.M().Info("清理 Redis 资源")
	}

	if p.worker != nil {
		p.worker.Release()
		logger.M().Info("停止 Worker")
	}

	if p.watcher != nil {
		p.watcher.Close()
		logger.M().Info("停止配置监听器")
	}
}

func (p *Rain) initWorker() (err error) {
	ants.Release()

	if p.config.Worker.Capacity == 0 {
		p.config.Worker.Capacity = 1000
	}

	p.worker, err = ants.NewPool(p.config.Worker.Capacity, p.config.Worker.Options()...)
	if err != nil {
		return
	}

	logger.M().Info("初始化 Worker")

	return nil
}

func (p *Rain) OnStart(callback func()) {
	p.beforeStart = append(p.beforeStart, callback)
}

func (p *Rain) OnStop(callback func()) {
	p.beforeStop = append(p.beforeStop, callback)
}

func (p *Rain) OnConfigUpdate(callback func(config *Config)) {
	p.onConfigUpdate = callback
}

func (p *Rain) Router() *gin.Engine {
	return p.engine
}
