package compatibility

import "testing"

func TestCalculateCompatibilityPercent(t *testing.T) {
	tests := []struct {
		name     string
		earnedAB int
		totalAB  int
		earnedBA int
		totalBA  int
		want     int
	}{
		{
			name:     "perfect match",
			earnedAB: 10,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  20,
			want:     100,
		},
		{
			name:     "no match",
			earnedAB: 0,
			totalAB:  10,
			earnedBA: 0,
			totalBA:  20,
			want:     0,
		},
		{
			name:     "one sided half match",
			earnedAB: 5,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  20,
			want:     71,
		},
		{
			name:     "zero totals default to full satisfaction",
			earnedAB: 0,
			totalAB:  0,
			earnedBA: 0,
			totalBA:  0,
			want:     100,
		},
		{
			name:     "clamps above one hundred",
			earnedAB: 20,
			totalAB:  10,
			earnedBA: 20,
			totalBA:  10,
			want:     100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCompatibilityPercent(tt.earnedAB, tt.totalAB, tt.earnedBA, tt.totalBA)
			if got != tt.want {
				t.Fatalf("calculateCompatibilityPercent() = %d, want %d", got, tt.want)
			}
		})
	}
}
