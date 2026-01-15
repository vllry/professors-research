package prizeodds

import "testing"

func TestOddsTable_Get(t *testing.T) {
	// Test some known values
	tests := []struct {
		x    int
		y    int
		want float64
	}{
		{0, 0, 1.0},   // If 0 copies in deck, 100% chance of prizing 0
		{0, 1, 0.0},   // If 0 copies in deck, 0% chance of prizing 1
		{1, 0, 0.9},   // If 1 copy in deck, 90% chance of prizing 0
		{1, 1, 0.1},   // If 1 copy in deck, 10% chance of prizing 1
		{60, 0, 0.0},  // Out of range x
		{0, 7, 0.0},   // Out of range y
	}

	for _, tt := range tests {
		got := DefaultOddsTable.Get(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("DefaultOddsTable.Get(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestOddsTable_Get_SumToOne(t *testing.T) {
	// For any x, the sum of probabilities for y=0..6 should be approximately 1.0
	for x := 0; x <= 59; x++ {
		sum := 0.0
		for y := 0; y <= 6; y++ {
			sum += DefaultOddsTable.Get(x, y)
		}
		// Allow small floating point error
		if sum < 0.9999 || sum > 1.0001 {
			t.Errorf("Sum of probabilities for x=%d is %v, expected ~1.0", x, sum)
		}
	}
}






