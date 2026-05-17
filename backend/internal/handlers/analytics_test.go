package handlers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePearson(t *testing.T) {
	tests := []struct {
		name     string
		x        []float64
		y        []float64
		expected float64
	}{
		{
			name:     "Perfect Positive Correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{2, 4, 6, 8, 10},
			expected: 1.0,
		},
		{
			name:     "Perfect Negative Correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{-1, -2, -3, -4, -5},
			expected: -1.0,
		},
		{
			name:     "Zero Correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{5, 1, 5, 1, 5},
			expected: 0.0,
		},
		{
			name:     "No Variance X (Division by Zero prevention)",
			x:        []float64{1, 1, 1, 1, 1},
			y:        []float64{1, 2, 3, 4, 5},
			expected: 0.0,
		},
		{
			name:     "No Variance Y (Division by Zero prevention)",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{1, 1, 1, 1, 1},
			expected: 0.0,
		},
		{
			name:     "Small Dataset",
			x:        []float64{1, 10},
			y:        []float64{1, 10},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePearson(tt.x, tt.y)
			if math.IsNaN(result) {
				t.Errorf("calculatePearson() returned NaN for %v", tt.name)
			}
			// Use a small epsilon for float comparison
			if tt.expected == 0 {
				assert.InDelta(t, tt.expected, result, 0.0001)
			} else {
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}
