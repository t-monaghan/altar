package utils

// ResponseStatusIsNot2xx is a helper function to decide if a http response is OK.
func ResponseStatusIsNot2xx(httpStatus int) bool {
	return httpStatus < 200 || httpStatus > 299
}
