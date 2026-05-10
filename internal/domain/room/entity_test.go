package room

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusAvailable, true},
		{StatusOccupied, true},
		{StatusMaintenance, true},
		{StatusCleaning, true},
		{StatusReserved, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidStatus(tt.status))
		})
	}
}
