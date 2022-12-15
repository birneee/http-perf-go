package internal

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type HostnameDirectoryMultiplexHandler interface {
	http.Handler
	Hostnames() []string
}

type hostnameDirectoryMultiplexHandler struct {
	inner HostnameMultiplexHandler
}

func (h *hostnameDirectoryMultiplexHandler) Hostnames() []string {
	return h.inner.Hostnames()
}

func (h *hostnameDirectoryMultiplexHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.inner.ServeHTTP(writer, request)
}

func NewHostnameDirectoryMultiplexHandler(hostnameDirectory string, config FileServerConfig) (HostnameDirectoryMultiplexHandler, error) {
	entries, err := os.ReadDir(hostnameDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to list hostname directory: %v", err)
	}
	inner := NewHostnameMultiplexHandler()
	for _, entry := range entries {
		if entry.IsDir() {
			inner.AddHostname(entry.Name(), NewFileServer(http.Dir(filepath.Join(hostnameDirectory, entry.Name())), config))
		}
	}
	return &hostnameDirectoryMultiplexHandler{
		inner: inner,
	}, nil
}
