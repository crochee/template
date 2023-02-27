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
