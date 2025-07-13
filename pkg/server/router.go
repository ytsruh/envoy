package server

import "net/http"

// Router defines the methods for registering routes.
type Router interface {
	Get(pattern string, handler http.HandlerFunc)
	Post(pattern string, handler http.HandlerFunc)
	Put(pattern string, handler http.HandlerFunc)
	Patch(pattern string, handler http.HandlerFunc)
	Delete(pattern string, handler http.HandlerFunc)
	http.Handler
}

// router handles routing of HTTP requests.
type router struct {
	mux *http.ServeMux
}

// New creates a new Router.
func NewRouter() Router {
	return &router{mux: http.NewServeMux()}
}

// Get registers a new GET route.
func (r *router) Get(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("GET "+pattern, handler)
}

// Post registers a new POST route.
func (r *router) Post(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("POST "+pattern, handler)
}

// PUT registers a new PUT route.
func (r *router) Put(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("PUT "+pattern, handler)
}

// Patch registers a new PATCH route.
func (r *router) Patch(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("PATCH "+pattern, handler)
}

// Delete registers a new DELETE route.
func (r *router) Delete(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("DELETE "+pattern, handler)
}

// ServeHTTP makes the Router implement the http.Handler interface.
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
