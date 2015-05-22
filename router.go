package router

import (
	"net/http"
	"net/url"
	"regexp"
)

type Args []string
type Kwargs map[string]string
type RouteHandler func(http.ResponseWriter, *http.Request, Args, Kwargs)

type RouteRule struct {
	re      *regexp.Regexp
	handler RouteHandler
}

type Handler struct {
	rules []RouteRule
}

var default_handler *Handler = NewHandler()
var http404 http.Handler = http.NotFoundHandler()

func NewHandler() *Handler {
	ret := new(Handler)
	ret.rules = make([]RouteRule, 0)
	return ret
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, rule := range h.rules {
		if rule.Handle(w, req) {
			return
		}
	}
	http404.ServeHTTP(w, req)
}

func (h *Handler) Handle(regex string, handler RouteHandler) {
	nh := RouteRule{
		re:      regexp.MustCompile(regex),
		handler: handler,
	}
	h.rules = append(h.rules, nh)
}

func (h *Handler) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, h)
}

func (h *Handler) ListenAndServeTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, h)
}

func (r *RouteRule) Match(url string) bool {
	return r.re.MatchString(url)
}

func (r *RouteRule) MatchURL(url *url.URL) bool {
	if url == nil {
		return false
	}
	return r.re.MatchString(url.Path)
}

func (r *RouteRule) Handle(w http.ResponseWriter, req *http.Request) bool {
	if w == nil || req == nil {
		return false
	}
	subm := r.re.FindStringSubmatch(req.URL.Path)
	// Not matched.
	if subm == nil {
		return false
	}
	// Matched, build args and kwargs.
	args := Args(subm)
	kwargs := make(Kwargs)
	for i, names := 1, r.re.SubexpNames(); i <= r.re.NumSubexp(); i++ {
		if names[i] != "" {
			kwargs[names[i]] = args[i]
		}
	}
	r.handler(w, req, args, kwargs)
	return true
}
func Handle(regex string, handler RouteHandler) {
	default_handler.Handle(regex, handler)
}

func ListenAndServe(addr string) error {
	return default_handler.ListenAndServe(addr)
}

func ListenAndServeTLS(addr string, certFile string, keyFile string) error {
	return default_handler.ListenAndServeTLS(addr, certFile, keyFile)
}
