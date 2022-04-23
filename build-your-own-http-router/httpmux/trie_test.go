package httpmux

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

type testHandlerResponse struct {
	Desc   string `json:"desc"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

func testHandler(desc string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(testHandlerResponse{
			Desc:   desc,
			Path:   r.Method,
			Method: r.URL.Path,
		})
	})
}

func TestTrie_Insert(t *testing.T) {
	t.Run("simple routers", func(t *testing.T) {
		trie := NewTrie()

		ExpectErrNil(t, trie.Insert("/v1/users", "POST", testHandler("POST /v1/users")))
		ExpectErrNil(t, trie.Insert("/v1/users", "GET", testHandler("GET /v1/users")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles", "GET", testHandler("GET /v1/users/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles", "PATCH", testHandler("PATCH /v1/users/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles/settings", "POST", testHandler("POST /v1/users/profiles/settings")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles/settings", "GET", testHandler("GET /v1/users/profiles/settings")))

		userHandlers := trie.root.Children["v1"].Children["users"]

		ExpectMethodHandlerExists(t, userHandlers.Value, "POST")
		ExpectMethodHandlerExists(t, userHandlers.Value, "GET")
		ExpectMethodHandlerExists(t, userHandlers.Children["profiles"].Value, "GET")
		ExpectMethodHandlerExists(t, userHandlers.Children["profiles"].Value, "PATCH")
		ExpectMethodHandlerExists(t, userHandlers.Children["profiles"].Children["settings"].Value, "POST")
		ExpectMethodHandlerExists(t, userHandlers.Children["profiles"].Children["settings"].Value, "GET")
	})

	t.Run("vars routers", func(t *testing.T) {
		trie := NewTrie()

		ExpectErrNil(t, trie.Insert("/v1/users/{uid}", "GET", testHandler("GET /v1/users/{id}")))
		ExpectErrNil(t, trie.Insert("/v1/users/{uid}/profiles", "GET", testHandler("GET /v1/users/{id}/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/{uid}/profiles/{pid}", "GET", testHandler("GET /v1/users/{uid}/profiles/{pid}")))
		ExpectErrNil(t, trie.Insert("/v1/users/static/profiles/{pid}", "GET", testHandler("GET /v1/users/static/profiles/{pid}")))

		userHandlers := trie.root.Children["v1"].Children["users"]

		ExpectMethodHandlerExists(t, userHandlers.Children[VarsLabel].Value, "GET")
		ExpectTrue(t, userHandlers.Children[VarsLabel].Label == "uid")
		ExpectMethodHandlerExists(t, userHandlers.Children[VarsLabel].Children["profiles"].Value, "GET")
		ExpectTrue(t, userHandlers.Children[VarsLabel].Children["profiles"].Label == "profiles")
		ExpectMethodHandlerExists(t, userHandlers.Children[VarsLabel].Children["profiles"].Children[VarsLabel].Value, "GET")
		ExpectTrue(t, userHandlers.Children[VarsLabel].Children["profiles"].Children[VarsLabel].Label == "pid")
		ExpectMethodHandlerExists(t, userHandlers.Children["static"].Children["profiles"].Children[VarsLabel].Value, "GET")
		ExpectTrue(t, userHandlers.Children["static"].Children["profiles"].Children[VarsLabel].Label == "pid")
	})

}

func TestTrie_Get(t *testing.T) {
	t.Run("simple routers", func(t *testing.T) {
		trie := NewTrie()

		ExpectErrNil(t, trie.Insert("/v1/users", "POST", testHandler("POST /v1/users")))
		ExpectErrNil(t, trie.Insert("/v1/users", "GET", testHandler("GET /v1/users")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles", "GET", testHandler("GET /v1/users/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles", "PATCH", testHandler("PATCH /v1/users/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles/settings", "POST", testHandler("POST /v1/users/profiles/settings")))
		ExpectErrNil(t, trie.Insert("/v1/users/profiles/settings", "GET", testHandler("GET /v1/users/profiles/settings")))

		ExpectPathRegisteredWithVars(t, trie, "/v1/users", "POST", []Var{})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users", "GET", []Var{})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/profiles", "GET", []Var{})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/profiles", "PATCH", []Var{})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/profiles/settings", "POST", []Var{})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/profiles/settings", "GET", []Var{})
	})

	t.Run("vars routers", func(t *testing.T) {
		trie := NewTrie()

		ExpectErrNil(t, trie.Insert("/v1/users/{uid}", "GET", testHandler("GET /v1/users/{id}")))
		ExpectErrNil(t, trie.Insert("/v1/users/static", "GET", testHandler("GET /v1/users/static")))
		ExpectErrNil(t, trie.Insert("/v1/users/{uid}/profiles", "GET", testHandler("GET /v1/users/{id}/profiles")))
		ExpectErrNil(t, trie.Insert("/v1/users/{uid}/profiles/{pid}", "GET", testHandler("GET /v1/users/{uid}/profiles/{pid}")))
		ExpectErrNil(t, trie.Insert("/v1/users/static/profiles/{pid}", "GET", testHandler("GET /v1/users/static/profiles/{pid}")))

		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1", "GET", []Var{{Name: "uid", Value: "1"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1234", "GET", []Var{{Name: "uid", Value: "1234"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/abc", "GET", []Var{{Name: "uid", Value: "abc"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/static", "GET", []Var{})

		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1/profiles", "GET", []Var{{Name: "uid", Value: "1"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1234/profiles", "GET", []Var{{Name: "uid", Value: "1234"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/abc/profiles", "GET", []Var{{Name: "uid", Value: "abc"}})

		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1/profiles/1", "GET", []Var{{Name: "uid", Value: "1"}, {Name: "pid", Value: "1"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/1234/profiles/2", "GET", []Var{{Name: "uid", Value: "1234"}, {Name: "pid", Value: "2"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/abc/profiles/3", "GET", []Var{{Name: "uid", Value: "abc"}, {Name: "pid", Value: "3"}})
		ExpectPathRegisteredWithVars(t, trie, "/v1/users/static/profiles/3", "GET", []Var{{Name: "pid", Value: "3"}})

	})
}

func ExpectPathRegisteredWithVars(t *testing.T, trie *Trie, path string, method string, expVars []Var) {
	h, vars, err := trie.Get(path, method)
	if err != nil {
		t.Fatalf("expected error nil; got error: %v", err)
	}

	if !reflect.DeepEqual(expVars, vars) {
		t.Fatalf("expected %+v; got %+v", expVars, vars)
	}

	if h == nil {
		t.Fatalf("expcted handler exist")
	}
}

func ExpectTrue(t *testing.T, cond bool) {
	if !cond {
		t.Helper()
		t.Errorf("expecting true")
	}
}

func ExpectErrNil(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Errorf("expect error nil; got %v", err)
	}
}

func ExpectMethodHandlerExists(t *testing.T, values map[string]http.Handler, method string) {
	_, exists := values[method]
	if !exists {
		t.Helper()
		t.Errorf("expect method %s exists", method)
	}
}
