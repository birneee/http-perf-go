package internal

import (
	"fmt"
	"strings"
)

type Logger interface {
	Logf(level string, format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

type HierarchicalLogger interface {
	Logger
	Name() string
	Parent() HierarchicalLogger
	NewChild(name string) HierarchicalLogger
	// Breadcrumbs return list of ancestors starting from oldest
	Breadcrumbs() []string
}

type hierarchicalLogger struct {
	parent    HierarchicalLogger
	name      string
	formatter Formatter
}

func (l *hierarchicalLogger) Tracef(format string, args ...interface{}) {
	l.Logf("trace", format, args...)
}

func (l *hierarchicalLogger) Debugf(format string, args ...interface{}) {
	l.Logf("debug", format, args...)
}

func (l *hierarchicalLogger) Warnf(format string, args ...interface{}) {
	l.Logf("warn", format, args...)
}

func (l *hierarchicalLogger) Fatalf(format string, args ...interface{}) {
	l.Logf("fatal", format, args...)
}

func (l *hierarchicalLogger) Panicf(format string, args ...interface{}) {
	l.Logf("panic", format, args...)
}

var _ HierarchicalLogger = &hierarchicalLogger{}

// TODO move to own repo
func NewHierarchicalLogger(name string, formatter Formatter) HierarchicalLogger {
	h := &hierarchicalLogger{
		parent:    nil,
		name:      name,
		formatter: formatter,
	}
	return h
}

func (l *hierarchicalLogger) Infof(format string, args ...interface{}) {
	l.Logf("info", format, args...)
}

func (l *hierarchicalLogger) Errorf(format string, args ...interface{}) {
	l.Logf("error", format, args...)
}

func (l *hierarchicalLogger) Logf(level string, format string, args ...interface{}) {
	println(l.formatter.Format(l, level, fmt.Sprintf(format, args...)))
}
func (l *hierarchicalLogger) Name() string {
	return l.name
}

func (l *hierarchicalLogger) Parent() HierarchicalLogger {
	return l.parent
}

func (l *hierarchicalLogger) NewChild(name string) HierarchicalLogger {
	c := &hierarchicalLogger{
		parent:    l,
		name:      name,
		formatter: l.formatter,
	}
	return c
}

func (l *hierarchicalLogger) Breadcrumbs() []string {
	var names []string
	var iterator HierarchicalLogger = l
	for iterator != nil {
		name := iterator.Name()
		if name != "" {
			names = append([]string{name}, names...)
		}
		iterator = iterator.Parent()
	}
	return names
}

type Formatter interface {
	Format(logger HierarchicalLogger, level string, message string) string
}

type DefaultFormatter struct{}

func (d DefaultFormatter) Format(logger HierarchicalLogger, level string, message string) string {
	var parts = logger.Breadcrumbs()
	parts = append(parts, strings.ToUpper(level))

	return fmt.Sprintf("%s %s", strings.Join(parts, "|"), message)
}
