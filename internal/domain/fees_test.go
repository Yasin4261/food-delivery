package domain

import "testing"

func TestFeePolicy(t *testing.T) {
	p := FeePolicy{DeliveryBaseFee: 10, DeliveryFeePerKm: 2, CommissionRate: 15}

	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"delivery with distance", p.DeliveryFee(3.5), 17},        // 10 + 2*3.5
		{"delivery unknown distance", p.DeliveryFee(-1), 10},      // base only
		{"delivery zero distance", p.DeliveryFee(0), 10},          // same spot
		{"delivery rounds to cents", p.DeliveryFee(1.234), 12.47}, // 10 + 2.468
		{"commission", p.Commission(200), 30},
		{"commission rounds", p.Commission(9.99), 1.5}, // 1.4985 -> 1.50
		{"zero commission rate", FeePolicy{}.Commission(100), 0},
		{"free delivery policy", FeePolicy{}.DeliveryFee(5), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("= %v, want %v", tc.got, tc.want)
			}
		})
	}
}
