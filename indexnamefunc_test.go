package conv

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCaseInsensitiveIndexName(t *testing.T) {
	m := map[string]interface{}{
		"":    1,
		"a":   2,
		"b":   3,
		"bB":  4,
		"bBb": 5,
		"Cc":  6,
	}

	tests := []struct {
		key   string
		value interface{}
		ok    bool
	}{
		{"", 1, true},
		{"A", 2, true},
		{"b", 3, true},
		{"BbB", 5, true},
		{"cc", 6, true},
		{"d", nil, false},
		{"x", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, got1 := CaseInsensitiveIndexName(m, tt.key)
			if !reflect.DeepEqual(got, tt.value) {
				t.Errorf("CaseInsensitiveIndexName() got = %v, want %v", got, tt.value)
			}
			if got1 != tt.ok {
				t.Errorf("CaseInsensitiveIndexName() got1 = %v, want %v", got1, tt.ok)
			}
		})
	}
}

func TestCamelSnakeCaseIndexName(t *testing.T) {
	m := map[string]interface{}{
		"":    1,
		"_":   2,
		"__":  3,
		"A":   4,
		"a_b": 5,
		"a__": 6,
		"a b": 7,
	}

	tests := []struct {
		key       string
		wantValue interface{}
		wantOk    bool
	}{
		{"", 1, true},
		{"_", 2, true},
		{"__", 3, true},
		{"A", 4, true},
		{"a", 4, true},
		{"A_B", 5, true},
		{"a_B", 5, true},
		{"a_b", 5, true},
		{"a__", 6, true},
		{"A__", 6, true},
		{"a b", 7, true},

		{"AB", nil, false},
		{"_A", nil, false},
		{"__A", nil, false},
		{"__a", nil, false},
		{"A b", nil, false},
		{"x", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			gotValue, gotOk := CamelSnakeCaseIndexName(m, tt.key)
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("CamelSnakeCaseIndexName() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("CamelSnakeCaseIndexName() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_camelSnakeCaseCompare(t *testing.T) {
	tests := []struct {
		sx   string
		sy   string
		want bool
	}{
		{"", "", true},
		{"  ", "  ", true},
		{"", "1", false},
		{" ", "", false},
		{"_", "", false},
		{"_", "__", false},

		{"abc", "abc", true},
		{"abc", "Abc", true},
		{"aa_bb_cc", "aa_bb_cc", true},
		{"aa_bb_cc", "Aa_Bb_Cc", true},
		{"aa_bb_cc", "AA_Bb_Cc", false},
		{"aa_bb_cc", "aa__bb_cc", false},
		{"aa_bb_cc", "_aa_bb_cc", false},
		{"aa_bb_cc", "aa_bb_cc_", false},
		{"aa_bb_cc", "aa_bbcc", false},
		{"aa_bb_cc", "aabbcc", false},

		{"aa_bb_cc", "AaBbCc", true},
		{"aa_bb_cc", "AaBb_Cc", true},
		{"aa_bb_cc", "Aa_bbCc", true},
		{"aa_bb_cc", "aa_bBCc", false},
		{"aa_bb_cc", "aaBbcc", false},

		{"AaBbCc", "AaBbCc", true},
		{"AaBbCc", "AabbCc", false},

		// Heading or trailing underscores.
		{"_aa", "_aa", true},
		{"_aa", "_Aa", false},
		{"aa_", "aa", false},

		// With spaces.
		{" a_b c ", " a_b c ", true},
		{"AaBbCc", "Aa BbCc", false},
		{"a_b", "a_b ", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint("(", tt.sx, ":", tt.sy, ")"), func(t *testing.T) {
			if got := camelSnakeCaseCompare(tt.sx, tt.sy); got != tt.want {
				t.Errorf("camelSnakeCaseCompare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_camelSnakeNameIter(t *testing.T) {
	checkNext := func(t *testing.T, iter *camelSnakeNameIter, wantIdx int, wantCur rune, isWordStart bool) {
		iter.next()

		if iter.idx != wantIdx {
			t.Errorf("want idx=%v, got %v", wantIdx, iter.idx)
		}

		if iter.Current != wantCur {
			t.Errorf("idx=%v, want Current=%v, got %v", iter.idx, wantCur, iter.Current)
		}

		if iter.IsWordStart != isWordStart {
			t.Errorf("idx=%v, want IsWordStart=%v, got %v", iter.idx, isWordStart, iter.IsWordStart)
		}
	}

	t.Run("empty", func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune("")}
		checkNext(t, iter, -1, 0, false)
	})

	t.Run("space", func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune("  ")}
		checkNext(t, iter, 1, ' ', true)
		checkNext(t, iter, 2, ' ', false)
		checkNext(t, iter, -1, 0, false)
	})

	s := "_"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, '_', true)
		checkNext(t, iter, -1, 0, false)
	})

	s = "__"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, '_', true)
		checkNext(t, iter, 2, '_', false)
		checkNext(t, iter, -1, 0, false)
	})

	s = "_a"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, '_', true)
		checkNext(t, iter, 2, 'a', false)
		checkNext(t, iter, -1, 0, false)
	})

	s = "__a"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, '_', true)
		checkNext(t, iter, 2, '_', false)
		checkNext(t, iter, 3, 'a', false)
		checkNext(t, iter, -1, 0, false)
	})

	s = "ab_cd"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, 'a', true)
		checkNext(t, iter, 2, 'b', false)
		checkNext(t, iter, 4, 'c', true)
		checkNext(t, iter, 5, 'd', false)
		checkNext(t, iter, -1, 0, false)
	})

	s = "a_b_c_d"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, 'a', true)
		checkNext(t, iter, 3, 'b', true)
		checkNext(t, iter, 5, 'c', true)
		checkNext(t, iter, 7, 'd', true)
		checkNext(t, iter, -1, 0, false)
	})

	s = "a__b"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, 'a', true)
		checkNext(t, iter, 3, '_', true)
		checkNext(t, iter, 4, 'b', false)
		checkNext(t, iter, -1, 0, false)
	})

	s = "a___"
	t.Run(s, func(t *testing.T) {
		iter := &camelSnakeNameIter{s: []rune(s)}
		checkNext(t, iter, 1, 'a', true)
		checkNext(t, iter, 3, '_', true)
		checkNext(t, iter, 4, '_', false)
		checkNext(t, iter, -1, 0, false)
	})
}
