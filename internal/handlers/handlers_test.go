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
		{
			name: "test 4",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge",
				contentType: "text/plain",
			},
			want: want{
				code:     404,
				response: "",
			},
		},
		{
			name: "test 5",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/counter/testCounter/100",
				contentType: "text/plain",
			},
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name: "test 6",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/counter/",
				contentType: "text/plain",
			},
			want: want{
				code:     404,
				response: "",
			},
		},
		{
			name: "test 7",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/counter/testCounter/none",
				contentType: "text/plain",
			},
			want: want{
				code:     400,
				response: "",
			},
		},
		{
			name: "test 8",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge/testGauge/100",
				contentType: "text/plain",
			},
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name: "test 9",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge/",
				contentType: "text/plain",
			},
			want: want{
				code:     404,
				response: "",
			},
		},
		{
			name: "test 10",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/gauge/testGauge/none",
				contentType: "text/plain",
			},
			want: want{
				code:     400,
				response: "",
			},
		},
		{
			name: "test 11",
			sendParam: sendParam{
				httpMethod:  http.MethodPost,
				request:     "/update/unknown/testCounter/100",
				contentType: "text/plain",
			},
			want: want{
				code:     501,
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

			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
