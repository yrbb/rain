package rain

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/yrbb/rain/pkg/logger"
	"github.com/yrbb/rain/pkg/middleware"
)

func registerServerCommand(p *Rain) {
	p.cmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "http server",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			p.isServer = true

			if err := p.config.Server.validate(); err != nil {
				return err
			}

			p.engine.Use(
				middleware.Logger(),
				middleware.Recovery(),
			)

			if p.config.Debug || p.config.Server.EnablePProf {
				pprof.Register(p.engine)
			}

			p.engine.GET("/health", func(c *gin.Context) {
				_, _ = c.Writer.WriteString("ok")
			})

			p.server = &http.Server{
				Addr:    p.config.Server.Listen,
				Handler: p.engine,
			}

			if p.config.Server.ReadTimeout > 0 {
				p.server.ReadTimeout = p.config.Server.ReadTimeout
			}

			if p.config.Server.WriteTimeout > 0 {
				p.server.WriteTimeout = p.config.Server.WriteTimeout
			}

			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			logger.M().Info(fmt.Sprintf("HTTP 服务启动, 监听: %s", p.config.Server.Listen))

			if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}

			return nil
		},
	})
}

func serverIsReady(listen string) bool {
	if strings.HasPrefix(listen, "0.0.0.0") {
		listen = strings.Replace(listen, "0.0.0.0", "127.0.0.1", 1)
	}

	res, err := http.Get(fmt.Sprintf("http://%s/health", listen))
	if err != nil {
		return false
	}

	_ = res.Body.Close()

	return res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound
}
