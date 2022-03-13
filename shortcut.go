package conv

import (
	"reflect"
	"time"
)

// Provides a group of shortcut methods for convenient use, to avoid initializing the Conv struct.

var defaultConv = new(Conv)

// ConvertType is equivalent to new(Conv).ConvertType() .
func ConvertType(src interface{}, dstTyp reflect.Type) (interface{}, error) {
	return defaultConv.ConvertType(src, dstTyp)
}

// Convert is equivalent to new(Conv).Convert() .
func Convert(src interface{}, dstPtr interface{}) error {
	return defaultConv.Convert(src, dstPtr)
}

// Bool converts the given value to the corresponding value of bool.
// The value must be simple, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).SimpleToBool(v) .
func Bool(v interface{}) (bool, error) {
	return defaultConv.SimpleToBool(v)
}

// String converts the given value to the corresponding value of string.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).SimpleToString(v) .
func String(v interface{}) (string, error) {
	return defaultConv.SimpleToString(v)
}

// Int converts the given value to the corresponding value of int.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(int(0))) .
func Int(v interface{}) (int, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int)
	if err != nil {
		return 0, err
	}
	return res.(int), nil
}

// Int64 converts the given value to the corresponding value of int64.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(int64(0))) .
func Int64(v interface{}) (int64, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int64)
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

// Int32 converts the given value to the corresponding value of int32.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(int32(0))) .
func Int32(v interface{}) (int32, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int32)
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}

// Int16 converts the given value to the corresponding value of int16.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(int16(0))) .
func Int16(v interface{}) (int16, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int16)
	if err != nil {
		return 0, err
	}
	return res.(int16), nil
}

// Int8 converts the given value to the corresponding value of int8.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(int8(0))) .
func Int8(v interface{}) (int8, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int8)
	if err != nil {
		return 0, err
	}
	return res.(int8), nil
}

// Uint converts the given value to the corresponding value of uint.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(uint(0))) .
func Uint(v interface{}) (uint, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint)
	if err != nil {
		return 0, err
	}
	return res.(uint), nil
}

// Uint64 converts the given value to the corresponding value of uint64.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(uint(0))) .
func Uint64(v interface{}) (uint64, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint64)
	if err != nil {
		return 0, err
	}
	return res.(uint64), nil
}

// Uint32 converts the given value to the corresponding value of uint32.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(uint(0))) .
func Uint32(v interface{}) (uint32, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint32)
	if err != nil {
		return 0, err
	}
	return res.(uint32), nil
}

// Uint16 converts the given value to the corresponding value of uint16.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(uint(0))) .
func Uint16(v interface{}) (uint16, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint16)
	if err != nil {
		return 0, err
	}
	return res.(uint16), nil
}

// Uint8 converts the given value to the corresponding value of uint8.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(uint(0))) .
func Uint8(v interface{}) (uint8, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint8)
	if err != nil {
		return 0, err
	}
	return res.(uint8), nil
}

// Float64 converts the given value to the corresponding value of float64.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(float64(0))) .
func Float64(v interface{}) (float64, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Float64)
	if err != nil {
		return 0, err
	}
	return res.(float64), nil
}

// Float32 converts the given value to the corresponding value of float32.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(float32(0))) .
func Float32(v interface{}) (float32, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Float32)
	if err != nil {
		return 0, err
	}
	return res.(float32), nil
}

// Complex128 converts the given value to the corresponding value of complex128.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(complex128(0+0i))) .
func Complex128(v interface{}) (complex128, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Complex128)
	if err != nil {
		return 0, err
	}
	return res.(complex128), nil
}

// Complex64 converts the given value to the corresponding value of complex64.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).Convert(v, reflect.TypeOf(complex64(0+0i))) .
func Complex64(v interface{}) (complex64, error) {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Complex64)
	if err != nil {
		return 0, err
	}
	return res.(complex64), nil
}

// Time converts the given value to the corresponding value of time.Time.
// The value must be a simple type, for which IsSimpleType() returns true.
// It is equivalent to new(Conv).SimpleToSimple(v, reflect.TypeOf(time.Time{})) .
func Time(v interface{}) (time.Time, error) {
	res, err := defaultConv.SimpleToSimple(v, typTime)
	if err != nil {
		return zeroTime, err
	}
	return res.(time.Time), nil
}

// MapToStruct is equivalent to new(Conv).MapToStruct() .
func MapToStruct(m map[string]interface{}, dstTyp reflect.Type) (interface{}, error) {
	return defaultConv.MapToStruct(m, dstTyp)
}

// StructToMap is equivalent to new(Conv).StructToMap() .
func StructToMap(v interface{}) (map[string]interface{}, error) {
	return defaultConv.StructToMap(v)
}

// MustConvertType is equivalent to new(Conv).MustConvertType() .
func MustConvertType(src interface{}, dstTyp reflect.Type) interface{} {
	return defaultConv.MustConvertType(src, dstTyp)
}

// MustConvert is equivalent to new(Conv).MustConvert() .
func MustConvert(src interface{}, dstPtr interface{}) {
	defaultConv.MustConvert(src, dstPtr)
}

// MustBool is like Bool() but panics instead of returns an error.
func MustBool(v interface{}) bool {
	res, err := defaultConv.SimpleToBool(v)
	if err != nil {
		panic(err)
	}
	return res
}

// MustString is like String() but panics instead of returns an error.
func MustString(v interface{}) string {
	res, err := defaultConv.SimpleToString(v)
	if err != nil {
		panic(err)
	}
	return res
}

// MustInt is like Int() but panics instead of returns an error.
func MustInt(v interface{}) int {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int)
	if err != nil {
		panic(err)
	}
	return res.(int)
}

// MustInt64 is like Int64() but panics instead of returns an error.
func MustInt64(v interface{}) int64 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int64)
	if err != nil {
		panic(err)
	}
	return res.(int64)
}

// MustInt32 is like Int32() but panics instead of returns an error.
func MustInt32(v interface{}) int32 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int32)
	if err != nil {
		panic(err)
	}
	return res.(int32)
}

// MustInt16 is like Int16() but panics instead of returns an error.
func MustInt16(v interface{}) int16 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int16)
	if err != nil {
		panic(err)
	}
	return res.(int16)
}

// MustInt8 is like Int8() but panics instead of returns an error.
func MustInt8(v interface{}) int8 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Int8)
	if err != nil {
		panic(err)
	}
	return res.(int8)
}

// MustUint is like Uint() but panics instead of returns an error.
func MustUint(v interface{}) uint {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint)
	if err != nil {
		panic(err)
	}
	return res.(uint)
}

// MustUint64 is like Uint64() but panics instead of returns an error.
func MustUint64(v interface{}) uint64 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint64)
	if err != nil {
		panic(err)
	}
	return res.(uint64)
}

// MustUint32 is like Uint32() but panics instead of returns an error.
func MustUint32(v interface{}) uint32 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint32)
	if err != nil {
		panic(err)
	}
	return res.(uint32)
}

// MustUint16 is like Uint16() but panics instead of returns an error.
func MustUint16(v interface{}) uint16 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint16)
	if err != nil {
		panic(err)
	}
	return res.(uint16)
}

// MustUint8 is like Uint8() but panics instead of returns an error.
func MustUint8(v interface{}) uint8 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Uint8)
	if err != nil {
		panic(err)
	}
	return res.(uint8)
}

// MustFloat64 is like Float64() but panics instead of returns an error.
func MustFloat64(v interface{}) float64 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Float64)
	if err != nil {
		panic(err)
	}
	return res.(float64)
}

// MustFloat32 is like Float32() but panics instead of returns an error.
func MustFloat32(v interface{}) float32 {
	res, err := defaultConv.simpleToPrimitive(v, reflect.Float32)
	if err != nil {
		panic(err)
	}
	return res.(float32)
}

// MustMapToStruct is like MapToStruct() but panics instead of returns an error.
func MustMapToStruct(m map[string]interface{}, dstTyp reflect.Type) interface{} {
	res, err := defaultConv.MapToStruct(m, dstTyp)
	if err != nil {
		panic(err)
	}
	return res
}

// MustStructToMap is like StructToMap() but panics instead of returns an error.
func MustStructToMap(v interface{}) map[string]interface{} {
	res, err := defaultConv.StructToMap(v)
	if err != nil {
		panic(err)
	}
	return res
}
