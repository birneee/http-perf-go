package common

import (
	"bufio"
	"fmt"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
	"io"
	"os"
)

func NewQlogTracer(filePrefix string, onFileCreate func(filename string)) logging.Tracer {
	return qlog.NewTracer(func(p logging.Perspective, connectionID []byte) io.WriteCloser {
		filename := fmt.Sprintf("%s_%x.qlog", filePrefix, connectionID)
		f, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		if onFileCreate != nil {
			onFileCreate(filename)
		}
		return NewBufferedWriteCloser(bufio.NewWriter(f), f)
	})
}
