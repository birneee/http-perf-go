package internal

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type HostnameMultiplexHandler interface {
	http.Handler
	AddHostname(hostname string, handler http.Handler)
	Hostnames() []string
}

type hostnameMultiplexHandler struct {
	handlers map[string]http.Handler
}

func (h *hostnameMultiplexHandler) Hostnames() []string {
	hostnames := make([]string, 0, len(h.handlers))
	for hostname := range h.handlers {
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

func (h *hostnameMultiplexHandler) AddHostname(hostname string, handler http.Handler) {
	h.handlers[hostname] = handler
}

func (h *hostnameMultiplexHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	hostname := request.Host
	if strings.Contains(hostname, ":") {
		var err error
		hostname, _, err = net.SplitHostPort(request.Host)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Printf("failed to multiplex hostname: %v\n", err)
			return
		}
	}
	handler, ok := h.handlers[hostname]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	handler.ServeHTTP(writer, request)
	writer.Header().Set("Content-Security-Policy", "default-src *") //TODO make this configurable
}

func NewHostnameMultiplexHandler() HostnameMultiplexHandler {
	return &hostnameMultiplexHandler{
		handlers: map[string]http.Handler{},
	}
}
