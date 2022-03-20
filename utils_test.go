package conv

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIsPrimitiveKind(t *testing.T) {
	tests := []struct {
		args reflect.Kind
		want bool
	}{
		{reflect.Bool, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Int, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Uint, true},
		{reflect.Float32, true},
		{reflect.Float64, true},
		{reflect.Complex64, true},
		{reflect.Complex128, true},
		{reflect.String, true},
		{reflect.Array, false},
		{reflect.Slice, false},
		{reflect.Ptr, false},
		{reflect.Uintptr, false},
		{reflect.Map, false},
		{reflect.Chan, false},
	}

	for _, tt := range tests {
		name := fmt.Sprint(tt.args)
		t.Run(name, func(t *testing.T) {
			if got := IsPrimitiveKind(tt.args); got != tt.want {
				t.Errorf("IsPrimitiveKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSimpleType(t *testing.T) {
	type args struct {
		t reflect.Type
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil", args{reflect.TypeOf(nil)}, false},
		{"true", args{reflect.TypeOf(true)}, true},
		{"time", args{reflect.TypeOf(time.Now())}, true},
		{"number", args{reflect.TypeOf(123)}, true},
		{"struct", args{reflect.TypeOf(struct{}{})}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSimpleType(tt.args.t); got != tt.want {
				t.Errorf("IsSimpleType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_errCantConvertTo(t *testing.T) {
	e := errCantConvertTo(99, "dst")
	want := "cannot convert 99 (int) to dst"
	if e.Error() != want {
		t.Errorf("got %#v, want %#v", e.Error(), want)
	}
}

func Test_errValueOverflow(t *testing.T) {
	e := errValueOverflow(true, "dst")
	want := "value overflow when converting true (bool) to dst"
	if e.Error() != want {
		t.Errorf("got %#v, want %#v", e.Error(), want)
	}
}

func Test_errPrecisionLoss(t *testing.T) {
	e := errPrecisionLoss(1.5, "dst")
	want := "lost precision when converting 1.5 (float64) to dst"
	if e.Error() != want {
		t.Errorf("got %#v, want %#v", e.Error(), want)
	}
}

func Test_getFieldPath(t *testing.T) {
	type a struct{ X, Y int }
	type B struct{ a }
	type c struct{ *B }
	type D struct{ c }

	t.Run("D", func(t *testing.T) {
		typ := reflect.TypeOf(D{})
		path := getFieldPath(typ, []int{0, 0, 0, 1})
		if path != "c.B.a.Y" {
			t.Fatal(path)
		}
	})

	t.Run("pB", func(t *testing.T) {
		typ := reflect.TypeOf(&B{})
		path := getFieldPath(typ, []int{0, 0})
		if path != "a.X" {
			t.Fatal(path)
		}
	})

	t.Run("a", func(t *testing.T) {
		typ := reflect.TypeOf(a{})
		path := getFieldPath(typ, []int{1})
		if path != "Y" {
			t.Fatal(path)
		}
	})
}

func Test_getFieldValue(t *testing.T) {
	type a struct {
		X int
		Y *int
	}

	t.Run("not-embedded", func(t *testing.T) {
		y := 33
		val := reflect.ValueOf(&a{X: 22, Y: &y}).Elem()

		res, err := getFieldValue(val, []int{0})
		if err != nil {
			t.Fatalf("get X: %v", err)
		}

		if res.Interface() != 22 {
			t.Error("X!=22")
		}

		res, err = getFieldValue(val, []int{1})
		if err != nil {
			t.Fatalf("get Y: %v", err)
		}

		if *res.Interface().(*int) != 33 {
			t.Error("Y!=33")
		}
	})

	t.Run("err", func(t *testing.T) {
		type B struct{ *a }

		// Get B.a.Y, try to init B.a which cannot be set.
		val := reflect.ValueOf(&B{})
		_, err := getFieldValue(val, []int{0, 1})
		if err == nil {
			t.Fatal("a: need error")
		}
	})

	t.Run("panic", func(t *testing.T) {
		p := func(msg string, fn func()) {
			defer func() {
				r := recover()
				if r != msg {
					t.Fatalf("want '%s', got '%v'", msg, r)
				}
			}()

			fn()
		}

		var pa *struct{}
		i := 1
		p("index must be given", func() { getFieldValue(reflect.ValueOf(&a{}), []int{}) })
		p("value is nil", func() { getFieldValue(reflect.ValueOf(&pa), []int{0}) })
		p("value must be struct", func() { getFieldValue(reflect.ValueOf(&i), []int{0}) })
	})

	t.Run("embedded", func(t *testing.T) {
		type B struct{ a }
		type c struct{ *B } // *B should be initialized.
		type D struct{ *c }

		val := reflect.ValueOf(&D{c: &c{}})
		res, err := getFieldValue(val, []int{0, 0, 0, 0}) // D.c.B.a.X
		if err != nil {
			t.Fatal(err)
		}

		if res.Interface() != 0 {
			t.Error("X!=0")
		}

		res, err = getFieldValue(val, []int{0, 0, 0, 1}) // D.c.B.a.X
		if err != nil {
			t.Fatal(err)
		}

		if res.Interface().(*int) != nil {
			t.Error("Y!=nil")
		}
	})
}
