package pprof

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	dg "runtime/debug"
	pf "runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/yrbb/rain/pkg/utils"
)

func New(cfg *Config) *Debug {
	return &Debug{config: cfg, startTime: time.Now()}
}

type Config struct {
	CPUFile   string `toml:"cpu_file"`
	HeapFile  string `toml:"heap_file"`
	TraceFile string `toml:"trace_file"`
}

type Debug struct {
	config    *Config
	startTime time.Time

	debugCf *os.File
	debugHf *os.File
	debugTf *os.File

	debugStats bool
}

func (d *Debug) Stats() {
	if !d.debugStats {
		d.StartStats()
	}

	d.StopStats()
}

func (d *Debug) StartStats() {
	for d.debugStats {
		elapsed := time.Since(d.startTime)

		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)

		slog.Debug(fmt.Sprintf(
			"alloc: %d, sys: %d, alloc(rate): %.2f",
			mem.Alloc,
			mem.Sys,
			float64(mem.TotalAlloc)/elapsed.Seconds(),
		))

		gc := dg.GCStats{PauseQuantiles: make([]time.Duration, 100)}
		if dg.ReadGCStats(&gc); gc.NumGC > 0 {
			slog.Debug(fmt.Sprintf(
				"numgc: %d, pause: %v, pause(avg): %v, overhead: %.2f, histogram: %v",
				gc.NumGC,
				gc.Pause[0],
				utils.AvgTime(gc.Pause),
				float64(gc.PauseTotal)/float64(elapsed)*100,
				[]any{gc.PauseQuantiles[94], gc.PauseQuantiles[98], gc.PauseQuantiles[99]},
			))
		}

		time.Sleep(time.Second * 5)
	}
}

func (d *Debug) StopStats() {
	d.debugStats = false
}

func (d *Debug) PProf() {
	if d.debugCf == nil {
		d.StartParseProfile()
		return
	}

	d.StopParseProfile()
}

func (d *Debug) StartParseProfile() {
	if d.debugCf != nil {
		return
	}

	cf := utils.If(d.config.CPUFile == "", "/tmp/cpu.prof", d.config.CPUFile)
	hf := utils.If(d.config.HeapFile == "", "/tmp/heap.prof", d.config.HeapFile)
	tf := utils.If(d.config.TraceFile == "", "/tmp/trace.out", d.config.TraceFile)

	d.debugCf, _ = os.Create(cf)
	d.debugHf, _ = os.Create(hf)
	d.debugTf, _ = os.Create(tf)

	pf.StartCPUProfile(d.debugCf)
	trace.Start(d.debugTf)
}

func (d *Debug) StopParseProfile() {
	if d.debugCf == nil {
		return
	}

	pf.StopCPUProfile()
	pf.WriteHeapProfile(d.debugHf)
	trace.Stop()

	d.debugCf.Close()
	d.debugHf.Close()
	d.debugTf.Close()

	d.debugCf = nil
	d.debugHf = nil
	d.debugTf = nil
}

func (d *Debug) DumpStacks() string {
	buf := make([]byte, 1024000)
	buf = buf[:runtime.Stack(buf, true)]

	return fmt.Sprintf("=== BEGIN goraintine stack dump ===\n%s\n=== END goraintine stack dump ===", buf)
}
