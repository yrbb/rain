package writer

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var _ Writer = &LevelWriter{}

func NewLevelWriter(project, path string, splitTime time.Duration) Writer {
	w := &LevelWriter{
		project:   project,
		path:      strings.TrimSuffix(path, "/") + "/",
		splitTime: splitTime,

		data:   map[slog.Level][]byte{},
		exitCh: make(chan struct{}),
		stopCh: make(chan struct{}),
	}

	go w.write()

	return w
}

type LevelWriter struct {
	m sync.Mutex
	w sync.WaitGroup

	project     string
	path        string
	splitTime   time.Duration
	closed      bool
	ignoreLevel bool

	files  sync.Map
	data   map[slog.Level][]byte
	exitCh chan struct{}
	stopCh chan struct{}
}

func (w *LevelWriter) write() {
	stop := false
	for !stop {
		w.m.Lock()
		s := w.data
		w.data = map[slog.Level][]byte{}
		w.m.Unlock()

		w.writeData(s)

		select {
		case <-w.stopCh:
			stop = true
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}

	w.writeData(w.data)

	close(w.exitCh)
}

func (w *LevelWriter) writeData(data map[slog.Level][]byte) {
	for l, v := range data {
		if len(v) == 0 {
			continue
		}

		w.w.Add(1)
		go func(l slog.Level, v []byte) {
			w.writeFile(l, v)
			w.w.Done()
		}(l, v)
	}
}

func (w *LevelWriter) writeFile(lvl slog.Level, data []byte) {
	h, err := w.handler(lvl)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if _, err = h.Write(data); err != nil {
		log.Println(err.Error())
	}
}

func (w *LevelWriter) Write(lvl slog.Level, data []byte) error {
	if w.ignoreLevel {
		lvl = slog.LevelInfo
	}

	w.m.Lock()
	w.data[lvl] = append(w.data[lvl], data...)
	w.m.Unlock()

	return nil
}

func (w *LevelWriter) handler(lvl slog.Level) (fp *os.File, err error) {
	now := time.Now().Unix()
	if w.splitTime > 0 {
		now -= now % int64(time.Minute*w.splitTime/time.Second)
	}

	ls := strings.ToLower(lvl.String())

	format := w.project + ".%s.log" // time.log
	args := []any{time.Unix(now, 0).Format("200601021504")}
	if !w.ignoreLevel {
		format = w.project + ".%s.%s.log" // level.time.log
		args = append([]any{ls}, args...)
	}

	file := fmt.Sprintf(format, args...)

	if tmp, ok := w.files.Load(ls); ok {
		if fp = tmp.(*os.File); strings.HasSuffix(fp.Name(), file) {
			return
		}

		_ = fp.Close()
	}

	fp, err = os.OpenFile(w.path+file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return
	}

	w.files.Store(ls, fp)

	return
}

func (w *LevelWriter) WithSplitTime(st time.Duration) {
	w.splitTime = st
}

func (w *LevelWriter) Close() {
	if w.closed {
		return
	}

	w.closed = true

	close(w.stopCh)
	<-w.exitCh
	w.w.Wait()

	w.files.Range(func(_, v any) bool {
		v.(*os.File).Close()
		return true
	})
}
