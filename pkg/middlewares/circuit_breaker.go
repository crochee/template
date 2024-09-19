package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vulcand/oxy/cbreaker"

	"template/pkg/code"
	"template/pkg/json"
	"template/pkg/resp"
)

// CircuitBreaker circuit breaker
func CircuitBreaker(
	ctx context.Context,
	next http.Handler,
) http.Handler {
	expression := "NetworkErrorRatio() > 0.8"
	cbOpts := []cbreaker.CircuitBreakerOption{
		cbreaker.Fallback(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusInternalServerError)
			data, err := json.Marshal(&resp.Response{
				Code: fmt.Sprintf(
					"%3d%s",
					code.ErrInternalServerError.StatusCode(),
					code.ErrInternalServerError.Code(),
				),
				Message: code.ErrInternalServerError.Message(),
				Result:  fmt.Sprintf("blocked by circuit-breaker (%q)", expression),
			})
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := rw.Write(data); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		})),
	}
	oxyCircuitBreaker, err := cbreaker.New(next, expression, cbOpts...)
	if err != nil {
		panic(err)
	}
	return oxyCircuitBreaker
}
