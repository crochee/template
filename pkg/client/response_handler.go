package client

import (
	"net/http"
)

type Parser struct {
	Handler func(*http.Response, interface{}) error
}

func (p Parser) Parse(resp *http.Response, result interface{}, opts ...func(*http.Response) error) error {
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return p.Handler(resp, result)
}
