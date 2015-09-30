package mstree

import (
	"io"
	"regexp"
	"strings"
	"sync"
)

type node struct {
	Children map[string]*node
	Lock     *sync.Mutex
}

const (
	TOKEN_MAX_LENGTH = 500
)

var (
	VALID_TOKEN_RE = regexp.MustCompile("^[a-z0-9A-Z_-]+$")
)

func newNode() *node {
	return &node{make(map[string]*node), new(sync.Mutex)}
}

func (n *node) insert(tokens []string, inserted *bool) {
	if len(tokens) == 0 {
		return
	}
	n.Lock.Lock()
	defer n.Lock.Unlock()

	first, tail := tokens[0], tokens[1:]
	if len(first) > TOKEN_MAX_LENGTH {
		log.Error("Token '%s' is too long, ignoring", first)
		return
	}

	if !VALID_TOKEN_RE.MatchString(first) {
		*inserted = false
		log.Error("Invalid token '%s' received, ignoring", first)
		return
	}

	child, ok := n.Children[first]
	if !ok {
		*inserted = true
		child = newNode()
		n.Children[first] = child
	}
	child.insert(tail, inserted)
}

func (n *node) TraverseDump(prefix string, writer io.Writer) {
	if len(n.Children) == 0 {
		io.WriteString(writer, prefix+"\n")
	} else {
		for k, node := range n.Children {
			var nPref string
			if prefix == "" {
				nPref = k
			} else {
				nPref = prefix + "." + k
			}
			node.TraverseDump(nPref, writer)
		}
	}
}

func (n *node) search(pattern string) map[string]*node {
	if pattern == "*" {
		return n.Children
	}

	results := make(map[string]*node)

	wcIndex := strings.Index(pattern, "*")
	qIndex := strings.Index(pattern, "?")
	obIndex := strings.Index(pattern, "[")
	cbIndex := strings.Index(pattern, "]")

	if wcIndex == -1 && qIndex == -1 && obIndex == -1 && cbIndex == -1 {
		if node, ok := n.Children[pattern]; ok {
			results[pattern] = node
		}
		return results
	}

	if cbIndex == -1 && obIndex == -1 {
		if qIndex == -1 {
			// Only *
			lwcIndex := strings.LastIndex(pattern, "*")

			if wcIndex != lwcIndex || (wcIndex != 0 && wcIndex != len(pattern)-1) {
				// more than one wildcard or one wildcard in the middle
				rePattern := "^" + strings.Replace(pattern, "*", ".*", -1) + "$"
				re, err := regexp.Compile(rePattern)
				if err != nil {
					return results
				}
				for k, node := range n.Children {
					if re.MatchString(k) {
						results[k] = node
					}
				}
				return results
			}

			if wcIndex == len(pattern)-1 {
				// wildcard at the end
				partial := pattern[:len(pattern)-1]
				for k, node := range n.Children {
					if strings.HasPrefix(k, partial) {
						results[k] = node
					}
				}
			} else {
				// wildcard at the begining
				partial := pattern[1:]
				for k, node := range n.Children {
					if strings.HasSuffix(k, partial) {
						results[k] = node
					}
				}
			}
		} else if wcIndex == -1 {
			// Only ?
			lqIndex := strings.LastIndex(pattern, "?")
			if qIndex != lqIndex || (qIndex != 0 && qIndex != len(pattern)-1) {
				// more than one ? or one ? in the middle
				rePattern := "^" + strings.Replace(pattern, "?", ".", -1) + "$"
				re, err := regexp.Compile(rePattern)
				if err != nil {
					return results
				}
				for k, node := range n.Children {
					if re.MatchString(k) {
						results[k] = node
					}
				}
				return results
			}

			if qIndex == len(pattern)-1 {
				// ? at the end
				partial := pattern[:len(pattern)-1]
				for k, node := range n.Children {
					if k[:len(k)-1] == partial {
						results[k] = node
					}
				}
			} else {
				// ? at the begining
				partial := pattern[1:]
				for k, node := range n.Children {
					if k[1:] == partial {
						results[k] = node
					}
				}
			}

		} else {
			// * and ? presents
			rePattern := "^" + strings.Replace(strings.Replace(pattern, "*", ".*", -1), "?", ".?", -1) + "$"
			re, err := regexp.Compile(rePattern)
			if err != nil {
				return results
			}
			for k, node := range n.Children {
				if re.MatchString(k) {
					results[k] = node
				}
			}
		}
	} else {
		rePattern := "^" + strings.Replace(strings.Replace(pattern, "*", ".*", -1), "?", ".?", -1) + "$"
		re, err := regexp.Compile(rePattern)
		if err != nil {
			return results
		}
		for k, node := range n.Children {
			if re.MatchString(k) {
				results[k] = node
			}
		}
	}

	return results
}
