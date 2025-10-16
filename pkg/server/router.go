package server

import "net/http"

// router handles routing of HTTP requests.
type Router struct {
	mux *http.ServeMux
}

// New creates a new Router.
func NewRouter() Router {
	return Router{mux: http.NewServeMux()}
}

// Get registers a new GET route.
func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("GET "+pattern, handler)
}

// Post registers a new POST route.
func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("POST "+pattern, handler)
}

// PUT registers a new PUT route.
func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("PUT "+pattern, handler)
}

// Patch registers a new PATCH route.
func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("PATCH "+pattern, handler)
}

// Delete registers a new DELETE route.
func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc("DELETE "+pattern, handler)
}

// ServeHTTP makes the Router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
