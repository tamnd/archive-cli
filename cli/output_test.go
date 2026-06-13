package cli

import "testing"

func TestScalarString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"hi", "hi"},
		{true, "true"},
		{float64(90), "90"},
		{float64(1.5), "1.5"},
		{[]any{"a", "b"}, "a; b"},
	}
	for _, c := range cases {
		if got := scalarString(c.in); got != c.want {
			t.Errorf("scalarString(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRowFieldStringReachesValueMap(t *testing.T) {
	// A field not in the curated columns must still be projectable from the
	// row's underlying value, so --fields never drops parsed data.
	v := map[string]any{"length": "02:01", "bitrate": float64(112)}
	if s, ok := rowFieldString(v, "length"); !ok || s != "02:01" {
		t.Errorf("length lookup = %q, %v", s, ok)
	}
	if s, ok := rowFieldString(v, "bitrate"); !ok || s != "112" {
		t.Errorf("bitrate lookup = %q, %v", s, ok)
	}
	if _, ok := rowFieldString(v, "missing"); ok {
		t.Error("missing key should report not-found")
	}
}

func TestProjectFallsBackToValue(t *testing.T) {
	o := &Output{fields: []string{"name", "length"}}
	r := Row{
		Cols:  []string{"name", "size"},
		Vals:  []string{"a.mp3", "1024"},
		Value: map[string]any{"name": "a.mp3", "size": "1024", "length": "02:01"},
	}
	cols, vals := o.project(r)
	if len(cols) != 2 || cols[1] != "length" {
		t.Fatalf("cols = %v", cols)
	}
	// "name" comes from the curated column; "length" only exists in Value.
	if vals[0] != "a.mp3" || vals[1] != "02:01" {
		t.Errorf("vals = %v, want [a.mp3 02:01]", vals)
	}
}
