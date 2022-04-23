package httpmux

import (
	"context"
	"net/http"
)

type contextType struct{}

var varsContextKey = &contextType{}

func contextWithVars(ctx context.Context, vars Vars) context.Context {
	return context.WithValue(ctx, varsContextKey, vars)
}

func GetVars(ctx context.Context) Vars {
	vars, _ := ctx.Value(varsContextKey).(Vars)
	return vars
}

type Router struct {
	trie *Trie
}

func NewRouter() *Router {
	return &Router{
		trie: NewTrie(),
	}
}

func (r *Router) Handle(method string, path string, handler http.Handler) {
	if err := r.trie.Insert(path, method, handler); err != nil {
		panic(err)
	}
}

func (r *Router) HandleFunc(method string, path string, handler http.HandlerFunc) {
	if err := r.trie.Insert(path, method, handler); err != nil {
		panic(err)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, vars, err := r.trie.Get(req.URL.Path, req.Method)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	ctx := contextWithVars(req.Context(), vars)
	handler.ServeHTTP(w, req.WithContext(ctx))
}
