package internal

import (
	"fmt"
	"strings"
	"time"
)

type formatter struct {
	start time.Time
}

var _ Formatter = &formatter{}

func NewFormatter() Formatter {
	return &formatter{
		start: time.Now(),
	}
}

func (f formatter) Format(logger HierarchicalLogger, level string, message string) string {
	parts := logger.Breadcrumbs()
	parts = append(parts, strings.ToUpper(level))
	return fmt.Sprintf("%s T%f %s", strings.Join(parts, "|"), time.Now().Sub(f.start).Seconds(), message)
}
