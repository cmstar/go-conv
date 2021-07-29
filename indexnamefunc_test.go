package conv

import (
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
