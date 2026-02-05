package conv

import (
	"reflect"
	"testing"
)

func TestFieldWalker_WalkFields(t *testing.T) {
	type want struct {
		Name  string
		Path  string
		Index []int
		Tag   string
	}

	check := func(t *testing.T, walker *FieldWalker, ws []want) {
		i := 0
		walker.WalkFields(func(f FieldInfo) bool {
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

			if f.TagValue != w.Tag {
				t.Fatalf("want tag %s, got %s", w.Tag, f.TagValue)
			}

			i++
			return true
		})

		if i != len(ws) {
			t.Fatalf("not enough fields")
		}
	}

	t.Run("empty", func(t *testing.T) {
		walker := NewFieldWalker(reflect.TypeOf(struct{}{}), "")
		count := 0
		walker.WalkFields(func(fi FieldInfo) bool {
			count++
			return true
		})

		if count != 0 {
			t.FailNow()
		}
	})

	t.Run("top2", func(t *testing.T) {
		type s struct{ X, Y, Z int }
		walker := NewFieldWalker(reflect.TypeOf(s{}), "")
		count := 0
		walker.WalkFields(func(fi FieldInfo) bool {
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
			B  int // hidden by T.B
			Ec     // hidden by T.Ec.D
			C  int
		}
		type T struct {
			A  int
			Eb `conv:"X"` // the tag will not be processed
			B  string     // hides Eb.B
			Ec
		}
		walker := NewFieldWalker(reflect.TypeOf(T{}), "")
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
			X int // hidden by T.B
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
		walker := NewFieldWalker(reflect.TypeOf(T{}), "c")
		check(t, walker, []want{
			{"B", "B", []int{2}, "X"},
			{"A", "A.A", []int{0, 0}, ""},
		})
	})

	t.Run("deep", func(t *testing.T) {
		type A struct {
			A1, A2 int
		}
		type B struct {
			B1, B2 int
			A
		}
		type C struct {
			C1, C2 int
			B
		}
		type D struct {
			D1, D2 int
			C
		}
		type E struct {
			E1, E2 int
			D
		}

		walker := NewFieldWalker(reflect.TypeOf(E{}), "")
		check(t, walker, []want{
			{"E1", "E1", []int{0}, ""},
			{"E2", "E2", []int{1}, ""},
			{"D1", "D.D1", []int{2, 0}, ""},
			{"D2", "D.D2", []int{2, 1}, ""},
			{"C1", "D.C.C1", []int{2, 2, 0}, ""},
			{"C2", "D.C.C2", []int{2, 2, 1}, ""},
			{"B1", "D.C.B.B1", []int{2, 2, 2, 0}, ""},
			{"B2", "D.C.B.B2", []int{2, 2, 2, 1}, ""},
			{"A1", "D.C.B.A.A1", []int{2, 2, 2, 2, 0}, ""},
			{"A2", "D.C.B.A.A2", []int{2, 2, 2, 2, 1}, ""},
		})
	})
}

func TestFieldWalker_WalkValues(t *testing.T) {
	type want struct {
		Value int // 0 if the field isn't int.
		Path  string
	}

	check := func(t *testing.T, walker *FieldWalker, val reflect.Value, ws []want) {
		i := 0
		walker.WalkValues(val, func(f FieldInfo, v reflect.Value) bool {
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
		walker := NewFieldWalker(reflect.TypeOf(a), "")
		walker.WalkValues(reflect.ValueOf(a), func(fi FieldInfo, v reflect.Value) bool {
			t.FailNow()
			return true
		})
	})

	t.Run("top1", func(t *testing.T) {
		var a struct{ X, Y int }
		walker := NewFieldWalker(reflect.TypeOf(a), "")

		count := 0
		walker.WalkValues(reflect.ValueOf(a), func(fi FieldInfo, v reflect.Value) bool {
			count++
			return count < 1
		})

		if count != 1 {
			t.FailNow()
		}
	})

	t.Run("hiding", func(t *testing.T) {
		type A struct {
			A1 int
			A2 int
			x  int
		}
		type B struct {
			A  // hidden by C.A
			B1 int
			A1 int // hidden by C.A.A1
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

		walker := NewFieldWalker(reflect.TypeOf(c), "")
		check(t, walker, reflect.ValueOf(c), []want{
			{1, "C1"},
			{2, "C2"},
			{0, "D"},
			{10, "B.B1"},
		})
	})

	t.Run("embedding", func(t *testing.T) {
		type A struct {
			A1 int
			A2 *int
		}
		type B struct {
			B1 int
			B2 *int
		}
		type C struct {
			C1 int
			C2 *int
		}
		type D struct {
			A
			*B
			*C
			D1 int
		}

		d := D{}
		d.A1 = 3
		d.C = &C{
			C1: 11,
		}
		d.D1 = 33

		walker := NewFieldWalker(reflect.TypeOf(d), "")
		check(t, walker, reflect.ValueOf(d), []want{
			{33, "D1"},
			{3, "A.A1"},
			{0, "A.A2"},
			{11, "C.C1"},
			{0, "C.C2"},
		})
	})
}
