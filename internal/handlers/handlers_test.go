package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func TestUpdateMetricHandler(t *testing.T) {
	var f *os.File = storage.OpenMetricFileCSV()
	defer f.Close()

	type sendParam struct {
		httpMethod  string
		request     string
		contentType string
	}

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name      string
		sendParam sendParam
		want      want
	}{
		{
			name: "test 1",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge/Alloc/0",
				contentType: "text/plain",
			},
			want: want{
				code:     200,
				response: `{"status":"ok"}`,
			},
		},
		{
			name: "test 2",
			sendParam: sendParam{
				httpMethod:  http.MethodGet,
				request:     "/update/gauge/Alloc/0",
				contentType: "text/plain",
			},
			want: want{
				code:     405,
				response: "",
			},
		},
		{
			name: "test 3",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge/Alloc/0",
				contentType: "text/html",
			},
			want: want{
				code:     415,
				response: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.sendParam.httpMethod,
				tt.sendParam.request, nil)
			request.Header.Set("Content-Type", tt.sendParam.contentType)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricHandler(f))

			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, result.StatusCode, tt.want.code)
		})
	}
}
