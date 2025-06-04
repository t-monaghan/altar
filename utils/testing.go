// Package utils is a package for general utilities for developing Altar
package utils

import "net/http"

// MockRoundTripper is a function type that implements the http.RoundTripper interface.
// It allows for mocking HTTP requests in tests.
type MockRoundTripper func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface for MockRoundTripper.
// It simply calls the underlying function with the provided request.
func (m MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

// MockClient returns an http.Client with a mocked transport function.
// This is useful for testing HTTP interactions without making actual network requests.
func MockClient(mockTransportFn func(r *http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: MockRoundTripper(mockTransportFn)}
}
