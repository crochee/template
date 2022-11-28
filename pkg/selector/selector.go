package selector

import (
	"net/url"
)

type Node struct {
	Name   string
	URL    url.URL
	Weight float64
}

type Selector interface {
	Next() *Node
}

type DcsSelector struct {
	Node
}

func (d *DcsSelector) Next() *Node {
	return &d.Node
}
