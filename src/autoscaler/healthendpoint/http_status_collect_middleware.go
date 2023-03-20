package healthendpoint

import (
	"fmt"
	"net/http"
	"os"
)

type HTTPStatusCollectMiddleware struct {
	httpStatusCollector HTTPStatusCollector
}

func NewHTTPStatusCollectMiddleware(httpStatusCollector HTTPStatusCollector) *HTTPStatusCollectMiddleware {
	return &HTTPStatusCollectMiddleware{
		httpStatusCollector: httpStatusCollector,
	}
}

func (h *HTTPStatusCollectMiddleware) Collect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.httpStatusCollector.IncConcurrentHTTPRequest()
		defer h.httpStatusCollector.DecConcurrentHTTPRequest()
		fmt.Fprintf(os.Stderr, "\n\nCHECKING w\n\n %v \n", w)
		fmt.Fprintf(os.Stderr, "\n\nCHECKING r\n\n %v \n", r)
		fmt.Fprintf(os.Stderr, "\n\nnext\n\n %v \n", next)
		next.ServeHTTP(w, r)
	})
}
