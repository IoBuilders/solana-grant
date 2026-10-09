package noderatelimit

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

// TooManyRequestsError is returned for an HTTP 429 from the node. JSON-RPC clients drop the response
// headers of a non-2xx answer, so the transport turns the 429 into this error before the client sees it,
// keeping the headers where a rate limit reset header can be read from.
type TooManyRequestsError struct {
	Header http.Header
	Body   []byte
}

func (e *TooManyRequestsError) Error() string {
	if len(e.Body) == 0 {
		return "node rate limit exceeded: 429 Too Many Requests"
	}
	return fmt.Sprintf("node rate limit exceeded: 429 Too Many Requests: %s", e.Body)
}

// transport wraps baseTransport so that every HTTP 429 from the nodes comes back as a TooManyRequestsError.
type transport struct {
	baseTransport http.RoundTripper
}

func NewTransport(baseTransport http.RoundTripper) http.RoundTripper {
	return &transport{baseTransport: baseTransport}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.baseTransport.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusTooManyRequests {
		return resp, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.ErrorWithCtx(req.Context(), "failed to close http response body", "error", err)
		}
	}(resp.Body)

	// A failed read only loses the body, which is informational; the 429 itself must still reach the
	// guard, so the error is logged rather than returned in place of the TooManyRequestsError.
	var body bytes.Buffer
	if _, err := body.ReadFrom(resp.Body); err != nil {
		logger.WarnWithCtx(req.Context(), "failed to read 429 response body from node", "error", err)
	}
	return nil, &TooManyRequestsError{Header: resp.Header.Clone(), Body: body.Bytes()}
}
