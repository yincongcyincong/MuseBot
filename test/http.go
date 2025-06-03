package test

import (
	"io"
	"net/http"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type MockBody struct {
	data []byte
	read int
}

func NewMockBody(data string) *MockBody {
	return &MockBody{data: []byte(data)}
}

func (m *MockBody) Read(p []byte) (n int, err error) {
	if m.read >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.read:])
	m.read += n
	return n, nil
}

func (m *MockBody) Close() error {
	return nil
}
