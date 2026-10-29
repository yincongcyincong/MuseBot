package llm

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"
)

type trackedAliyunBody struct {
	io.Reader
	closed bool
}

func (b *trackedAliyunBody) Close() error {
	b.closed = true
	return nil
}

func TestReadAliyunVideoResponseBodyClosesImmediately(t *testing.T) {
	readErr := errors.New("read failed")
	for _, tc := range []struct {
		name    string
		reader  io.Reader
		want    string
		wantErr error
	}{
		{name: "success", reader: strings.NewReader(`{"task_status":"RUNNING"}`), want: `{"task_status":"RUNNING"}`},
		{name: "read error", reader: iotest.ErrReader(readErr), wantErr: readErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedAliyunBody{Reader: tc.reader}
			data, err := readAliyunVideoResponseBody(&http.Response{Body: body})
			if !body.closed {
				t.Fatal("response body is still open after reading")
			}
			if !errors.Is(err, tc.wantErr) || string(data) != tc.want {
				t.Fatalf("data = %q, error = %v; want %q, %v", data, err, tc.want, tc.wantErr)
			}
		})
	}
}
