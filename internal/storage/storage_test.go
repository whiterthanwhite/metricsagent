package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenMetricFileCSV(t *testing.T) {
	tests := []struct {
		name         string
		isFileExists bool
	}{
		{
			name:         "test 1",
			isFileExists: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := OpenMetricFileCSV()
			f.Close()

			_, err := os.OpenFile("tmp.csv", os.O_RDWR, 0750)
			assert.Equal(t, tt.isFileExists, assert.Nil(t, err, ""))

			os.Remove("tmp.csv")
		})
	}
}
