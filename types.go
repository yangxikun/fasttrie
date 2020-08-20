package fasttrie

import (
	"regexp"
)

type nodeType uint8

type nodeWildcard struct {
	path     string
	paramKey string
	value    interface{}
}

type node struct {
	nType nodeType

	path         string
	tsr          bool
	value        interface{}
	hasWildChild bool
	children     []*node
	wildcard     *nodeWildcard

	paramKeys  []string
	paramRegex *regexp.Regexp
}

type wildPath struct {
	path  string
	keys  []string
	start int
	end   int
	pType nodeType

	pattern string
	regex   *regexp.Regexp
}

// Tree is a routes storage
type Tree struct {
	root *node

	// If enabled, the node value could be updated
	Mutable bool
}
