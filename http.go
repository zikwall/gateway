package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"
)

func httpWriteError(w http.ResponseWriter, err error) error {
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	return encoder.Encode(err)
}

type httpHandler struct {
	service      *description
	proxy        *httputil.ReverseProxy
	versionProxy *map[string]*httputil.ReverseProxy
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service.prefix != "" {
		r.URL.Path = strings.Replace(r.URL.Path, h.service.prefix, "", 1)
	}
	version := r.Header.Get("X-API-Version")
	if version != "" {
		if proxy, ok := (*h.versionProxy)[version]; ok {
			proxy.ServeHTTP(w, r)
			return
		}
	}
	h.proxy.ServeHTTP(w, r)
}

func newHTTPHandler(
	service *description,
	proxy *httputil.ReverseProxy,
	versionProxies map[string]*httputil.ReverseProxy,
) *httpHandler {
	return &httpHandler{
		service:      service,
		proxy:        proxy,
		versionProxy: &versionProxies,
	}
}
