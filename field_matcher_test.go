package conv

import (
	"reflect"
	"testing"
)

func TestSimpleMatcherCreator_caseInsensitive(t *testing.T) {
	type s struct {
		a, b, bB, Bb, cc, Ccc int
	}

	// Disable the warning from static-check.
	ss := s{}
	_, _, _, _, _, _ = ss.a, ss.b, ss.bB, ss.Bb, ss.cc, ss.Ccc

	ctor := SimpleMatcherCreator{
		Conf: SimpleMatcherConfig{
			CaseInsensitive: true,
		},
	}
	typ := reflect.TypeOf(s{})

	tests := []struct {
		name     string
		wantName string
		ok       bool
	}{
		{"", "", false},
		{"a", "", false},
		{"A", "", false},
		{"b", "", false},
		{"cc", "", false},

		{"bb", "Bb", true},
		{"BB", "Bb", true},
		{"cCC", "Ccc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mather := ctor.GetMatcher(typ)
			f, ok := mather.MatchField(tt.name)
			if f.Name != tt.wantName {
				t.Errorf("MatchField() name = %v, want %v", f.Name, tt.wantName)
			}
			if ok != tt.ok {
				t.Errorf("MatchField() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestSimpleMatcherCreator_omitUnderscore(t *testing.T) {
	type s struct {
		A_B_C int
	}

	// Disable the warning from static-check.
	ss := s{}
	_ = ss.A_B_C

	ctor := SimpleMatcherCreator{
		Conf: SimpleMatcherConfig{
			OmitUnderscore: true,
			CamelSnakeCase: true, // Should be ignored when OmitUnderscore is true.
		},
	}
	typ := reflect.TypeOf(s{})

	tests := []struct {
		name     string
		wantName string
		ok       bool
	}{
		{"", "", false},
		{"Abc", "", false},

		{"ABC", "A_B_C", true},
		{"A_BC", "A_B_C", true},
		{"A_B_C", "A_B_C", true},
		{"A_B_C_", "A_B_C", true},
		{"_A_B_C_", "A_B_C", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mather := ctor.GetMatcher(typ)
			f, ok := mather.MatchField(tt.name)
			if f.Name != tt.wantName {
				t.Errorf("MatchField() name = %v, want %v", f.Name, tt.wantName)
			}
			if ok != tt.ok {
				t.Errorf("MatchField() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestSimpleMatcherCreator_withTag(t *testing.T) {
	type s struct {
		A1 int `conv:"A"`
		A2 int `conv:"Bb"`
		A3 int `conv:"cc"`
	}

	// Disable the warning from static-check.
	ss := s{}
	_, _, _ = ss.A1, ss.A2, ss.A3

	ctor := SimpleMatcherCreator{
		Conf: SimpleMatcherConfig{
			Tag:             "conv",
			CaseInsensitive: true, // Can apply to tag values.
		},
	}
	typ := reflect.TypeOf(s{})

	tests := []struct {
		name     string
		wantName string
		ok       bool
	}{
		{"", "", false},
		{"A1", "", false},
		{"A2", "", false},
		{"A3", "", false},

		{"A", "A1", true},
		{"a", "A1", true}, // CaseInsensitive on.

		{"bb", "A2", true},
		{"Bb", "A2", true},
		{"cc", "A3", true},
		{"CC", "A3", true},
		{"CCc", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mather := ctor.GetMatcher(typ)
			f, ok := mather.MatchField(tt.name)
			if f.Name != tt.wantName {
				t.Errorf("MatchField() name = %v, want %v", f.Name, tt.wantName)
			}
			if ok != tt.ok {
				t.Errorf("MatchField() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestSimpleMatcherCreator_camelSnakeCase(t *testing.T) {
	type s struct {
		A, A__, Ab, A_b, A_B, A__B, AaBB, AaBBCc int
	}

	// Disable the warning from static-check.
	ss := s{}
	_, _, _, _, _, _, _, _ = ss.A, ss.A__, ss.Ab, ss.A_b, ss.A_B, ss.A__B, ss.AaBB, ss.AaBBCc

	ctor := SimpleMatcherCreator{
		Conf: SimpleMatcherConfig{
			CamelSnakeCase: true,
		},
	}
	typ := reflect.TypeOf(s{})

	tests := []struct {
		name     string
		wantName string
		ok       bool
	}{
		{"", "", false},
		{"_", "", false},
		{"__", "", false},
		{"_A", "", false},
		{"A_", "", false},
		{"_Aa_bb", "", false},
		{"aA_bb", "", false},

		{"a", "A", true},
		{"A", "A", true},
		{"A__", "A__", true},
		{"a__", "A__", true},

		{"Ab", "Ab", true},
		{"AB", "A_b", true},

		{"a_B", "A_b", true},
		{"aB", "A_b", true},
		{"A_b", "A_b", true},
		{"A_B", "A_b", true}, // The mather returns the first match A_b, hides A_B.

		{"A__B", "A__B", true},
		{"A__b", "", false},

		{"aaBB", "AaBB", true},
		{"AaBB", "AaBB", true},
		{"aa_BB", "AaBB", true}, // BB is treated as two words: B and B.
		{"aa_bb", "", false},    // bb is treated as one word.
		{"aa_Bb", "", false},
		{"aaBb", "", false},
		{"aabb", "", false},
		{"aA_BB", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mather := ctor.GetMatcher(typ)
			f, ok := mather.MatchField(tt.name)
			if f.Name != tt.wantName {
				t.Errorf("MatchField() name = %v, want %v", f.Name, tt.wantName)
			}
			if ok != tt.ok {
				t.Errorf("MatchField() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func Test_simpleMatcher_fixCamelSnakeCaseName(t *testing.T) {
	ix := &simpleMatcher{}

	tests := []struct {
		name string
		want string
	}{
		{"", ""},
		{"a", "_a"},
		{"_", "__"},
		{"__", "___"},
		{"___", "____"},
		{"____", "_____"},
		{"_a", "__a"},
		{"_A", "___a"},
		{"aa__bb", "_aa__bb"},  // aa + _bb
		{"Aa__Bb", "_aa___bb"}, // Aa + _ + Bb
		{"AaBb", "_aa_bb"},
		{"aaBb", "_aa_bb"},
		{"aa_bb_", "_aa_bb_"},   // aa + bb_
		{"aa_bb__", "_aa_bb__"}, // aa + bb__
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ix.fixCamelSnakeCaseName([]rune(tt.name)); got != tt.want {
				t.Errorf("fixCamelSnakeCaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_simpleMatcher_withEmbeddedStruct(t *testing.T) {
	checkLen := func(t *testing.T, m *syncMap, want int) bool {
		got := 0
		m.Range(func(k, v interface{}) bool {
			got++
			return true
		})
		if got != want {
			t.Errorf("want %d fields, got %d", want, got)
			return false
		}
		return true
	}

	checkValue := func(t *testing.T, m *syncMap, key string, wantType reflect.Type) bool {
		v, ok := m.Load(key)
		if !ok {
			t.Errorf("key %v does not exist", key)
			return false
		}

		got := v.(FieldInfo).Type
		if got != wantType {
			t.Errorf("key %v, want type %v, got %v", key, wantType, got)
			return false
		}

		return true
	}

	t.Run("embed", func(t *testing.T) {
		type A struct{ V int }
		type B struct{ A }
		type C struct{ B }

		mather := &simpleMatcher{typ: reflect.TypeOf(C{})}
		mather.initFieldMap()

		checkLen(t, mather.fs, 1)
		checkValue(t, mather.fs, "V", reflect.TypeOf(0))
	})

	t.Run("hide", func(t *testing.T) {
		type A struct {
			V1 int
			V2 int
			V3 int
		}
		type B struct {
			A
			V1 string // Hides A.V1.
		}
		type C struct {
			B
			V2 float64 // Hides A.V2.
		}

		mather := &simpleMatcher{typ: reflect.TypeOf(C{})}
		mather.initFieldMap()

		checkLen(t, mather.fs, 3)
		checkValue(t, mather.fs, "V1", reflect.TypeOf(""))
		checkValue(t, mather.fs, "V2", reflect.TypeOf(0.0))
		checkValue(t, mather.fs, "V3", reflect.TypeOf(0))
	})
}
