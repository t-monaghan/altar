package utils

import "errors"

// ErrClosingResponseBody describes the error encountered when a request fails whilst closing the response body.
var ErrClosingResponseBody = errors.New("failed to close response body")

// ResponseStatusIsNot2xx is a helper function to decide if a http response is OK.
func ResponseStatusIsNot2xx(httpStatus int) bool {
	return httpStatus < 200 || httpStatus > 299
}
