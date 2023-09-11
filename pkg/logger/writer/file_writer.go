package writer

import (
	"time"
)

func NewFileWriter(project, path string, splitTime time.Duration) Writer {
	w := NewLevelWriter(project, path, splitTime)
	w.(*LevelWriter).ignoreLevel = true

	return w
}
