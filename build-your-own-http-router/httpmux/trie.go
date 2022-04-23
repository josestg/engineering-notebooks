package httpmux

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

var varsNameRegex = regexp.MustCompile("^\\{[a-zA-z]+\\}$")

type NodeKind string

type Var struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Vars []Var

func (v Vars) ByName(name string) string {
	for _, e := range v {
		if e.Name == name {
			return e.Value
		}
	}

	return ""
}

const (
	RootNode = NodeKind("root")
	VarsNode = NodeKind("vars")
	PathNode = NodeKind("path")
)

const (
	RootLabel = "__ROOT__"
	VarsLabel = "__VARS__"
)

type TrieNode struct {
	Label    string
	Kind     NodeKind
	Value    map[string]http.Handler
	Children map[string]*TrieNode
}

func NewTrieNode(kind NodeKind, label string) *TrieNode {
	node := TrieNode{
		Kind:     kind,
		Label:    label,
		Value:    make(map[string]http.Handler),
		Children: make(map[string]*TrieNode),
	}

	return &node
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			Label:    RootLabel,
			Kind:     RootNode,
			Value:    make(map[string]http.Handler),
			Children: make(map[string]*TrieNode),
		},
	}
}

func (t *Trie) Insert(path string, method string, handler http.Handler) error {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	segments := strings.Split(path, "/")

	visitedNode := t.root

	for _, segment := range segments {
		nextNode, hasSegment := visitedNode.Children[segment]
		if hasSegment {
			visitedNode = nextNode
			continue
		}

		if varsNameRegex.MatchString(segment) {
			segment = strings.TrimSuffix(segment, "}")
			segment = strings.TrimPrefix(segment, "{")

			_, hasVars := visitedNode.Children[VarsLabel]
			if !hasVars {
				visitedNode.Children[VarsLabel] = NewTrieNode(VarsNode, segment)
			}

			visitedNode = visitedNode.Children[VarsLabel]
			if visitedNode.Label != segment {
				return errors.New("conflict location")
			}

			continue
		}

		visitedNode.Children[segment] = NewTrieNode(PathNode, segment)
		visitedNode = visitedNode.Children[segment]
	}

	_, hasMethodHandler := visitedNode.Value[method]
	if hasMethodHandler {
		return errors.New("conflict location")
	}

	visitedNode.Value[method] = handler
	return nil
}

func (t *Trie) Get(path string, method string) (http.Handler, Vars, error) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	segments := strings.Split(path, "/")

	vars := make([]Var, 0)
	visitedNode := t.root

	for _, segment := range segments {
		childNode, hasSegment := visitedNode.Children[segment]
		if !hasSegment {
			// try to check vars
			varsNode, hasVars := visitedNode.Children[VarsLabel]
			if !hasVars {
				return nil, vars, errors.New("handler not found")
			}

			vars = append(vars, Var{
				Name:  varsNode.Label,
				Value: segment,
			})

			visitedNode = varsNode
		} else {
			visitedNode = childNode
		}
	}

	handler, hasMethodHandler := visitedNode.Value[method]
	if !hasMethodHandler {
		return nil, vars, errors.New("handler not found")
	}

	return handler, vars, nil
}
