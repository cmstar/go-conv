package conv

import (
	"reflect"
	"testing"
	"time"
)

func TestConvertType(t *testing.T) {
	type args struct {
		src    interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"ok", args{"1", reflect.TypeOf(true)}, true, false},
		{"err", args{"err", reflect.TypeOf(true)}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertType(tt.args.src, tt.args.dstTyp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	var res int
	err := Convert("33", &res)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := 33
	if res != want {
		t.Errorf("want %v, got %v", want, res)
	}
}

func TestBool(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"ok", args{"true"}, true, false},
		{"err", args{"err"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Bool(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"ok", args{false}, "0", false},
		{"err", args{struct{}{}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"ok", args{"100"}, int(100), false},
		{"err", args{"err"}, int(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"ok", args{"100"}, int64(100), false},
		{"err", args{"err"}, int64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{"ok", args{"100"}, int32(100), false},
		{"err", args{"err"}, int32(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt16(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int16
		wantErr bool
	}{
		{"ok", args{"100"}, int16(100), false},
		{"err", args{"err"}, int16(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int16(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt8(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int8
		wantErr bool
	}{
		{"ok", args{"100"}, int8(100), false},
		{"err", args{"err"}, int8(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int8(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{"ok", args{"100"}, uint(100), false},
		{"err", args{"err"}, uint(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"ok", args{"100"}, uint64(100), false},
		{"err", args{"err"}, uint64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{"ok", args{"100"}, uint32(100), false},
		{"err", args{"err"}, uint32(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint16(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		wantErr bool
	}{
		{"ok", args{"100"}, uint16(100), false},
		{"err", args{"err"}, uint16(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint16(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint16() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint8(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    uint8
		wantErr bool
	}{
		{"ok", args{"100"}, uint8(100), false},
		{"err", args{"err"}, uint8(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint8(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"ok", args{"-33.5"}, float64(-33.5), false},
		{"err", args{"err"}, float64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Float64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Float64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Float64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat32(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float32
		wantErr bool
	}{
		{"ok", args{"-33.5"}, float32(-33.5), false},
		{"err", args{"err"}, float32(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Float32(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Float32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Float32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplex128(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    complex128
		wantErr bool
	}{
		{"ok", args{"-5+3i"}, complex128(-5 + 3i), false},
		{"err", args{"err"}, complex128(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Complex128(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complex128() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Complex128() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplex64(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    complex64
		wantErr bool
	}{
		{"ok", args{"-5+3i"}, complex64(-5 + 3i), false},
		{"err", args{"err"}, complex64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Complex64(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complex64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Complex64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{"ok", args{0}, time.Unix(0, 0), false},
		{"err", args{"err"}, zeroTime, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Time(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Time() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapToStruct(t *testing.T) {
	type args struct {
		m      map[string]interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"ok", args{map[string]interface{}{"I": 1}, reflect.TypeOf(S1{})}, S1{I: 1}, false},
		{"err", args{map[string]interface{}{"I": "err"}, reflect.TypeOf(S1{})}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapToStruct(tt.args.m, tt.args.dstTyp)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{"ok", args{S1{I: 1}}, map[string]interface{}{"I": 1, "F": 0.0, "S": ""}, false},
		{"err", args{S3{In: func() {}}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustConvertType(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustConvertType("1", reflect.TypeOf(1)) != 1 {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustConvertType(struct{}{}, reflect.TypeOf(1))
	})
}

func TestMustConvert(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var res int
		MustConvert("33", &res)
		if res != 33 {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		var res int
		MustConvert("g", &res)
	})
}

func TestMustBool(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustBool(1) != true {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustBool(struct{}{})
	})
}

func TestMustString(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustString(1) != "1" {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustString(struct{}{})
	})
}

func TestMustInt(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustInt("100") != 100 {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustInt(struct{}{})
	})
}

func TestMustInt64(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustInt64("100") != int64(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustInt64(struct{}{})
	})
}

func TestMustInt32(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustInt32("100") != int32(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustInt32(struct{}{})
	})
}

func TestMustInt16(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustInt16("100") != int16(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustInt16(struct{}{})
	})
}

func TestMustInt8(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustInt8("100") != int8(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustInt8(struct{}{})
	})
}

func TestMustUint(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustUint("100") != uint(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustUint(struct{}{})
	})
}

func TestMustUint64(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustUint64("100") != uint64(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustUint64(struct{}{})
	})
}

func TestMustUint32(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustUint32("100") != uint32(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustUint32(struct{}{})
	})
}

func TestMustUint16(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustUint16("100") != uint16(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustUint16(struct{}{})
	})
}

func TestMustUint8(t *testing.T) {

	t.Run("ok", func(t *testing.T) {
		if MustUint8("100") != uint8(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustUint8(struct{}{})
	})
}

func TestMustFloat64(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustFloat64("100") != float64(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustFloat64(struct{}{})
	})
}

func TestMustFloat32(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if MustFloat32("100") != float32(100) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustFloat32(struct{}{})
	})
}

func TestMustMapToStruct(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		res := MustMapToStruct(map[string]interface{}{"I": 1}, reflect.TypeOf(S1{}))
		if !reflect.DeepEqual(res, S1{I: 1}) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustMapToStruct(map[string]interface{}{"I": "err"}, reflect.TypeOf(S1{}))
	})
}

func TestMustStructToMap(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		res := MustStructToMap(S1{I: 1})
		if !reflect.DeepEqual(res, map[string]interface{}{"I": 1, "F": 0.0, "S": ""}) {
			t.FailNow()
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.FailNow()
			}
		}()

		MustStructToMap(1)
	})
}
