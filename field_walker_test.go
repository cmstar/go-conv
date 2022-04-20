package conv

import (
	"reflect"
	"testing"
)

func Test_fieldWalker_WalkFields(t *testing.T) {
	type want struct {
		Name  string
		Path  string
		Index []int
		Tag   string
	}

	check := func(t *testing.T, walker *fieldWalker, ws []want) {
		i := 0
		walker.WalkFields(func(f fieldInfo) bool {
			if i > len(ws)-1 {
				t.Fatalf("too many fields, index: %d", i)
			}

			w := ws[i]
			if f.Name != w.Name {
				t.Fatalf("want name %s, got %s", w.Name, f.Name)
			}

			if f.Path != w.Path {
				t.Fatalf("want path %s, got %s", w.Path, f.Path)
			}

			if !reflect.DeepEqual(f.Index, w.Index) {
				t.Fatalf("%s: want index %v, got %v", w.Name, w.Index, f.Index)
			}

			if f.Tag != w.Tag {
				t.Fatalf("want tag %s, got %s", w.Tag, f.Tag)
			}

			i++
			return true
		})

		if i != len(ws) {
			t.Fatalf("not enough fields")
		}
	}

	t.Run("empty", func(t *testing.T) {
		walker := newFieldWalker(reflect.TypeOf(struct{}{}), "")
		count := 0
		walker.WalkFields(func(fi fieldInfo) bool {
			count++
			return true
		})

		if count != 0 {
			t.FailNow()
		}
	})

	t.Run("top2", func(t *testing.T) {
		type s struct{ X, Y, Z int }
		walker := newFieldWalker(reflect.TypeOf(s{}), "")
		count := 0
		walker.WalkFields(func(fi fieldInfo) bool {
			count++
			return count < 2
		})

		if count != 2 {
			t.FailNow()
		}
	})

	t.Run("without-tag", func(t *testing.T) {
		type Ec struct {
			D int
			x int //lint:ignore U1000 Test unexported fields.
		}
		type Eb struct {
			B  int // hided by T.B
			Ec     // hided by T.Ec.D
			C  int
		}
		type T struct {
			A  int
			Eb `conv:"X"` // the tag will not be processed
			B  string     // hides Eb.B
			Ec
		}
		walker := newFieldWalker(reflect.TypeOf(T{}), "")
		check(t, walker, []want{
			{"A", "A", []int{0}, ""},
			{"B", "B", []int{2}, ""},
			{"C", "Eb.C", []int{1, 2}, ""},
			{"D", "Ec.D", []int{3, 0}, ""},
		})
	})

	t.Run("with-tag", func(t *testing.T) {
		type A struct {
			A int
			X int // hided by T.B
		}
		type B struct {
			B1 int // absent
			B2 int // absent
		}
		type T struct {
			*A
			s int     `c:"V"` //lint:ignore U1000 Test unexported fields.
			B `c:"X"` // hides B.X, the traverse will not go into the field
		}
		walker := newFieldWalker(reflect.TypeOf(T{}), "c")
		check(t, walker, []want{
			{"B", "B", []int{2}, "X"},
			{"A", "A.A", []int{0, 0}, ""},
		})
	})
}

func Test_fieldWalker_WalkValues(t *testing.T) {
	type want struct {
		Value int // 0 if the field isn't int.
		Path  string
	}

	check := func(t *testing.T, walker *fieldWalker, val reflect.Value, ws []want) {
		i := 0
		walker.WalkValues(val, func(f fieldInfo, v reflect.Value) bool {
			if i > len(ws)-1 {
				t.Errorf("too many fields, index: %d", i)
			}

			w := ws[i]
			if v.Kind() == reflect.Int && v.Interface() != w.Value {
				t.Errorf("index %d %s, want value %v, got %v", i, f.Name, w.Value, v.Interface())
			}

			if f.Path != w.Path {
				t.Errorf("want path %s, got %s", w.Path, f.Path)
			}

			i++
			return true
		})

		if i != len(ws) {
			t.Fatalf("not enough fields")
		}
	}

	t.Run("nil", func(t *testing.T) {
		var a *struct{}
		walker := newFieldWalker(reflect.TypeOf(a), "")
		walker.WalkValues(reflect.ValueOf(a), func(fi fieldInfo, v reflect.Value) bool {
			t.FailNow()
			return true
		})
	})

	t.Run("top1", func(t *testing.T) {
		var a struct{ X, Y int }
		walker := newFieldWalker(reflect.TypeOf(a), "")

		count := 0
		walker.WalkValues(reflect.ValueOf(a), func(fi fieldInfo, v reflect.Value) bool {
			count++
			return count < 1
		})

		if count != 1 {
			t.FailNow()
		}
	})

	t.Run("all", func(t *testing.T) {
		type A struct {
			A1 int
			A2 int
			x  int
		}
		type B struct {
			A  // hided by C.A
			B1 int
			A1 int // hided by C.A.A1
		}
		type C struct {
			C1 int
			C2 int
			*A // nil pointer will be ignored, hides B.A
			*B
			D struct{ X, Y int }
		}

		c := &C{
			C1: 1,
			C2: 2,
			B: &B{
				A:  A{100, 200, 300},
				B1: 10,
				A1: 20, // hides A.A1
			},
		}

		walker := newFieldWalker(reflect.TypeOf(c), "")
		check(t, walker, reflect.ValueOf(c), []want{
			{1, "C1"},
			{2, "C2"},
			{0, "D"},
			{10, "B.B1"},
		})
	})
}
