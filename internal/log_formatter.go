package internal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type formatter struct {
	start time.Time
}

func NewFormatter() log.Formatter {
	return &formatter{
		start: time.Now(),
	}
}

func (f *formatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s T%f %s\n", strings.ToUpper(entry.Level.String()), time.Now().Sub(f.start).Seconds(), entry.Message)), nil
}
