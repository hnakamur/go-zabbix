package zabbix

import "testing"

func TestParseAPIVersion(t *testing.T) {
	testCases := []string{
		"7.0.0alpha2",
		"6.4.4rc1",
		"6.4.3",
		"6.4.0beta6",
	}
	for _, c := range testCases {
		ver, err := ParseAPIVersion(c)
		if err != nil {
			t.Errorf("parse failed: input=%s, err=%v", c, err)
		}
		if got, want := ver.String(), c; got != want {
			t.Errorf("formatted version mismatch with input, got=%s, want=%s", got, want)
		}
	}
}

func TestAPIVersionCompare(t *testing.T) {
	testCases := []struct {
		v    string
		w    string
		want int
	}{
		{v: "7.0.0", w: "7.0.0alpha2", want: 1},
		{v: "7.0.0alpha1", w: "7.0.0alpha2", want: -1},
		{v: "7.0.0rc1", w: "7.0.0alpha2", want: 1},
		{v: "7.0.0rc1", w: "7.0.0beta2", want: 1},
		{v: "7.0.0rc1", w: "7.0.0rc1", want: 0},
		{v: "7.0.0", w: "7.0.0", want: 0},
		{v: "7.0.0rc1", w: "6.0.0rc1", want: 1},
		{v: "6.4.0rc1", w: "6.0.0rc1", want: 1},
		{v: "6.2.7", w: "6.0.13", want: 1},
		{v: "6.2.5", w: "6.2.6rc1", want: -1},
	}
	for _, c := range testCases {
		v, err := ParseAPIVersion(c.v)
		if err != nil {
			t.Errorf("parse failed: input=%s, err=%v", c.v, err)
		}
		w, err := ParseAPIVersion(c.w)
		if err != nil {
			t.Errorf("parse failed: input=%s, err=%v", c.w, err)
		}
		if got, want := v.Compare(w), c.want; got != want {
			t.Errorf("compare result mismatch, v=%s, want=%s, got=%d, want=%d", c.v, c.w, got, want)
		}
	}
}
