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
