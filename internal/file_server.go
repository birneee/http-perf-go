package internal

import (
	"net/http"
	"net/url"
	"strings"
)

type fileServer struct {
	inner  http.Handler
	config FileServerConfig
}

type FileServerConfig struct {
	QueryStringAsPartOfFile bool
}

// FileServer is http.Server with some additional options
type FileServer interface {
	http.Handler
}

func NewFileServer(root http.FileSystem, config FileServerConfig) FileServer {
	return &fileServer{
		inner:  http.FileServer(root),
		config: config,
	}
}

func (f fileServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if f.config.QueryStringAsPartOfFile {
		request.RequestURI = strings.ReplaceAll(request.RequestURI, "?", "%3F")
		var err error
		request.URL, err = url.ParseRequestURI(strings.ReplaceAll(request.URL.String(), "?", "%3F"))
		if err != nil {
			panic(err)
		}
	}
	f.inner.ServeHTTP(writer, request)
}
