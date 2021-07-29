package conv

import (
	"math"
	"testing"
)

func TestConv_primitiveToBool(t *testing.T) {
	tests := []struct {
		name    string
		args    interface{}
		want    bool
		wantErr bool
	}{
		{"0", 0, false, false},
		{"1", 1, true, false},
		{"-1", -1, true, false},
		{"-9999", -9999, true, false},
		{"55.3", 55.3, true, false},
		{"0+0i", 0 + 0i, false, false},
		{"0+1i", 0 + 1i, true, false},
		{"1+0i", 1 + 0i, true, false},

		{"err-emptye-string", "", false, true},
		{"err-wrong-string", "not-supported", false, true},
		{"err-struct", struct{}{}, false, true},
		{"err-slice", make([]struct{}, 0), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToBool(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToString(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"true", args{true}, "1"},
		{"false", args{false}, "0"},
		{"cn", args{"中"}, "中"},
		{"num", args{33}, "33"},
		{"complex0", args{complex128(0 + 0i)}, "0"},
		{"complex-123", args{complex64(-123 + 0i)}, "-123"},
		{"complex-123+99i", args{complex64(-123 + 99i)}, "(-123+99i)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Conv{}.primitiveToString(tt.args.v)
			if got != tt.want {
				t.Errorf("Conv.primitiveToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToInt64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"string", args{"9999"}, 9999, false},
		{"true", args{true}, 1, false},
		{"false", args{false}, 0, false},
		{"int", args{int(123456)}, 123456, false},
		{"int64", args{int64(123456)}, 123456, false},
		{"int32", args{int32(123456)}, 123456, false},
		{"int16", args{int16(12345)}, 12345, false},
		{"int8", args{int8(123)}, 123, false},
		{"uint", args{uint(123456)}, 123456, false},
		{"uint64", args{uint64(123456)}, 123456, false},
		{"uint32", args{uint32(123456)}, 123456, false},
		{"uint16", args{uint16(12345)}, 12345, false},
		{"uint8", args{uint8(123)}, 123, false},
		{"float64", args{float64(-876)}, -876, false},
		{"float32", args{float32(456)}, 456, false},
		{"complex64", args{complex64(5 + 0i)}, 5, false},
		{"complex128", args{complex64(-65560 + 0i)}, -65560, false},
		{"max", args{math.MaxInt64}, math.MaxInt64, false},
		{"min", args{math.MinInt64}, math.MinInt64, false},

		{"err-overflow-uint", args{uint64(math.MaxUint64)}, 0, true},
		{"err-overflow-float", args{float64(math.MaxUint64)}, 0, true},
		{"err-precision-loss1", args{1.5}, 0, true},
		{"err-precision-loss2", args{-0.1}, 0, true},
		{"err-imaginary-loss", args{-0.1 + 55i}, 0, true},
		{"err-string", args{"err"}, 0, true},
		{"err-struct", args{struct{}{}}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToInt64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToInt(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"err-string", args{"err"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToInt(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToInt32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{"max", args{math.MaxInt32}, math.MaxInt32, false},
		{"min", args{math.MinInt32}, math.MinInt32, false},
		{"err-overflow", args{math.MinInt64}, 0, true},
		{"err-string", args{"err"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToInt32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToInt16(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int16
		wantErr bool
	}{
		{"max", args{math.MaxInt16}, math.MaxInt16, false},
		{"min", args{math.MinInt16}, math.MinInt16, false},
		{"err-overflow", args{math.MinInt32}, 0, true},
		{"err-string", args{"err"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToInt16(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToInt16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToInt8(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int8
		wantErr bool
	}{
		{"max", args{math.MaxInt8}, math.MaxInt8, false},
		{"min", args{math.MinInt8}, math.MinInt8, false},
		{"err-overflow", args{math.MinInt16}, 0, true},
		{"err-string", args{"err"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToInt8(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToInt8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToInt8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToUint64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"string", args{"9999"}, uint64(9999), false},
		{"true", args{true}, uint64(1), false},
		{"false", args{false}, uint64(0), false},
		{"0", args{int(0)}, uint64(0), false},
		{"int", args{int(123456)}, uint64(123456), false},
		{"int64", args{int64(123456)}, uint64(123456), false},
		{"int32", args{int32(123456)}, uint64(123456), false},
		{"int16", args{int16(12345)}, uint64(12345), false},
		{"int8", args{int8(123)}, uint64(123), false},
		{"uint", args{uint(123456)}, uint64(123456), false},
		{"uint64", args{uint64(123456)}, uint64(123456), false},
		{"uint32", args{uint32(123456)}, uint64(123456), false},
		{"uint16", args{uint16(12345)}, uint64(12345), false},
		{"uint8", args{uint8(123)}, uint64(123), false},
		{"float64", args{float64(876)}, uint64(876), false},
		{"float32", args{float32(456)}, uint64(456), false},
		{"max", args{uint64(math.MaxUint64)}, uint64(math.MaxUint64), false},

		{"err-overflow-float", args{float64(math.MaxUint64) * 2}, uint64(0), true},
		{"err-precision-loss", args{1.5}, uint64(0), true},
		{"err-imaginary-loss", args{1 + 1i}, uint64(0), true},
		{"err-negative", args{-1}, uint64(0), true},
		{"err-string", args{"-1"}, uint64(0), true},
		{"err-struct", args{struct{}{}}, uint64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToUint64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToUint(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{"err-string", args{"err"}, uint(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToUint(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToUint32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{"max", args{uint32(math.MaxUint32)}, uint32(math.MaxUint32), false},
		{"err-overflow", args{uint64(math.MaxUint64)}, uint32(0), true},
		{"err-string", args{"err"}, uint32(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToUint32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToUint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToUint16(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		wantErr bool
	}{
		{"max", args{uint32(math.MaxUint16)}, uint16(math.MaxUint16), false},
		{"err-overflow", args{uint32(math.MaxUint32)}, uint16(0), true},
		{"err-string", args{"err"}, uint16(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToUint16(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToUint16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToUint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToUint8(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint8
		wantErr bool
	}{
		{"max", args{uint8(math.MaxUint8)}, uint8(math.MaxUint8), false},
		{"err-overflow", args{uint16(math.MaxUint16)}, uint8(0), true},
		{"err-string", args{"err"}, uint8(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToUint8(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToUint8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToUint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToFloat64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"string", args{"3.14"}, 3.14, false},
		{"true", args{true}, 1.0, false},
		{"false", args{false}, 0.0, false},
		{"int", args{int(-54321)}, -54321.0, false},
		{"int64", args{int64(-54321)}, -54321.0, false},
		{"int32", args{int32(-54321)}, -54321.0, false},
		{"int16", args{int16(-321)}, -321.0, false},
		{"int8", args{int8(-21)}, -21.0, false},
		{"uint", args{uint(54321)}, 54321.0, false},
		{"uint64", args{uint64(54321)}, 54321.0, false},
		{"uint32", args{uint32(54321)}, 54321.0, false},
		{"uint16", args{uint16(321)}, 321.0, false},
		{"uint8", args{uint8(21)}, 21.0, false},
		{"float32", args{float32(21.0)}, 21.0, false},
		{"float64", args{float64(21.0)}, 21.0, false},
		{"max", args{math.MaxFloat64}, math.MaxFloat64, false},

		{"err-imaginary-loss", args{-0.1 + 55i}, 0, true},
		{"err-string", args{"err"}, 0, true},
		{"err-map", args{map[string]int{}}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToFloat64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToFloat32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float32
		wantErr bool
	}{
		{"float32", args{math.MaxFloat32}, math.MaxFloat32, false},
		{"err-overflow", args{math.MaxFloat64}, 0, true},
		{"err-string", args{"err"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToFloat32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToFloat32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToFloat32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToComplex128(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    complex128
		wantErr bool
	}{
		{"string", args{"(33+5i)"}, 33 + 5i, false},
		{"true", args{true}, 1 + 0i, false},
		{"false", args{false}, 0 + 0i, false},
		{"int", args{int(123456)}, 123456 + 0i, false},
		{"int64", args{int64(123456)}, 123456 + 0i, false},
		{"int32", args{int32(123456)}, 123456 + 0i, false},
		{"int16", args{int16(12345)}, 12345 + 0i, false},
		{"int8", args{int8(123)}, 123 + 0i, false},
		{"uint", args{uint(123456)}, 123456 + 0i, false},
		{"uint64", args{uint64(123456)}, 123456 + 0i, false},
		{"uint32", args{uint32(123456)}, 123456 + 0i, false},
		{"uint16", args{uint16(12345)}, 12345 + 0i, false},
		{"uint8", args{uint8(123)}, 123 + 0i, false},
		{"float64", args{float64(-876.5)}, -876.5 + 0i, false},
		{"float32", args{float32(-876.5)}, -876.5 + 0i, false},

		{"err-string", args{"err"}, 0 + 0i, true},
		{"err-struct", args{struct{}{}}, 0 + 0i, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToComplex128(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToComplex128() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToComplex128() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_primitiveToComplex64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    complex64
		wantErr bool
	}{
		{"err", args{"err"}, 0 + 0i, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.primitiveToComplex64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Conv.primitiveToComplex64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Conv.primitiveToComplex64() = %v, want %v", got, tt.want)
			}
		})
	}
}
