package domain

import "testing"

// Stage times must always add up to the lesson length.
func TestKMJStageMinutes(t *testing.T) {
	cases := map[int][3]int{45: {10, 30, 5}, 40: {9, 27, 4}, 0: {10, 30, 5}}
	for d, want := range cases {
		s, m, e := KMJStageMinutes(d)
		if [3]int{s, m, e} != want {
			t.Errorf("KMJStageMinutes(%d) = %d/%d/%d, want %v", d, s, m, e, want)
		}
	}
	for d := KMJMinDuration; d <= KMJMaxDuration; d++ {
		s, m, e := KMJStageMinutes(d)
		if s+m+e != d || s < 1 || m < 1 || e < 1 {
			t.Errorf("KMJStageMinutes(%d) = %d/%d/%d does not add up", d, s, m, e)
		}
	}
}
