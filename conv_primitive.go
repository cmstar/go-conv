package conv

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Implements conversions between booleans, strings and numbers.
var primitive primitiveConv

type primitiveConv struct{}

func (c primitiveConv) toPrimitive(v interface{}, dstKind reflect.Kind) (interface{}, error) {
	switch dstKind {
	case reflect.Bool:
		return c.toBool(v)
	case reflect.String:
		return c.toString(v), nil
	case reflect.Int:
		return c.toInt(v)
	case reflect.Int8:
		return c.toInt8(v)
	case reflect.Int16:
		return c.toInt16(v)
	case reflect.Int32:
		return c.toInt32(v)
	case reflect.Int64:
		return c.toInt64(v)
	case reflect.Uint:
		return c.toUint(v)
	case reflect.Uint8:
		return c.toUint8(v)
	case reflect.Uint16:
		return c.toUint16(v)
	case reflect.Uint32:
		return c.toUint32(v)
	case reflect.Uint64:
		return c.toUint64(v)
	case reflect.Float32:
		return c.toFloat32(v)
	case reflect.Float64:
		return c.toFloat64(v)
	case reflect.Complex64:
		return c.toComplex64(v)
	case reflect.Complex128:
		return c.toComplex128(v)
	}

	// This should never run.
	panic("not a primitive type")
}

// toBool convert zero values to false, non-zero values to true.
func (c primitiveConv) toBool(v interface{}) (bool, error) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
	switch {
	case kind == reflect.String:
		return strconv.ParseBool(val.String())

	case kind == reflect.Bool:
		return val.Bool(), nil

	case isKindInt(kind):
		return val.Int() != 0, nil

	case isKindUint(kind):
		return val.Uint() != 0, nil

	case isKindFloat(kind):
		return val.Float() != 0, nil

	case isKindComplex(kind):
		return val.Complex() != 0, nil
	}

	return false, errCantConvertTo(v, "bool")
}

func (c primitiveConv) toString(v interface{}) string {
	switch vv := v.(type) {
	case bool:
		// The default string representation for booleans are true/false, which is not compatible
		// to other number types. To increase compatibility, we use 0/1 instead, they can be recognized
		// by strconv.ParseBool() , and can be converted to other number types.
		if vv {
			return "1"
		}
		return "0"

	case string:
		return vv

	case complex64:
		// Ignore the imaginary part of a complex number when it is zero, thus the value can be converted
		// to some other real number.
		// e.g., When converting (3+0i) to int, it is converted to "3" then converted to 3. If convert directly
		// from "(3+0i)" to int, it will result in an error.
		if imag(vv) == 0 {
			return fmt.Sprint(real(vv))
		}

	case complex128:
		if imag(vv) == 0 {
			return fmt.Sprint(real(vv))
		}
	}

	return fmt.Sprint(v)
}

func (c primitiveConv) doPrimitiveToInt64(v interface{}, dstType string) (int64, error) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
	switch {
	case kind == reflect.String:
		return strconv.ParseInt(val.String(), 0, 64)

	case kind == reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil

	case isKindInt(kind):
		return val.Int(), nil

	case isKindUint(kind):
		u := val.Uint()
		if u > math.MaxInt64 {
			return 0, errValueOverflow(v, dstType)
		}
		return int64(val.Uint()), nil

	case isKindFloat(kind):
		f := val.Float()
		return c.doFloat64ToInt64(f, dstType)

	case isKindComplex(kind):
		// Prevent data loss, ensure the imaginary part is zero.
		cpl := val.Complex()
		partImag := imag(cpl)
		if partImag != 0 {
			return 0, errImaginaryPartLoss(v, dstType)
		}

		partReal := real(cpl)
		return c.doFloat64ToInt64(partReal, dstType)
	}

	return 0, errCantConvertTo(v, dstType)
}

func (c primitiveConv) doFloat64ToInt64(f float64, dstType string) (int64, error) {
	if f < math.MinInt64 || f > math.MaxInt64 {
		return 0, errValueOverflow(f, dstType)
	}

	if f != math.Trunc(f) {
		return 0, errPrecisionLoss(f, dstType)
	}

	return int64(f), nil
}

func (c primitiveConv) toInt64(v interface{}) (int64, error) {
	return c.doPrimitiveToInt64(v, "int64")
}

func (c primitiveConv) toInt(v interface{}) (int, error) {
	num, err := c.doPrimitiveToInt64(v, "int")
	if err != nil {
		return 0, err
	}

	if num < minInt || num > maxInt {
		return 0, errValueOverflow(v, "int")
	}

	return int(num), nil
}

func (c primitiveConv) toInt32(v interface{}) (int32, error) {
	num, err := c.doPrimitiveToInt64(v, "int32")
	if err != nil {
		return 0, err
	}

	if num < math.MinInt32 || num > math.MaxInt32 {
		return 0, errValueOverflow(v, "int32")
	}

	return int32(num), nil
}

func (c primitiveConv) toInt16(v interface{}) (int16, error) {
	num, err := c.doPrimitiveToInt64(v, "int16")
	if err != nil {
		return 0, err
	}

	if num < math.MinInt16 || num > math.MaxInt16 {
		return 0, errValueOverflow(v, "int16")
	}

	return int16(num), nil
}

func (c primitiveConv) toInt8(v interface{}) (int8, error) {
	num, err := c.doPrimitiveToInt64(v, "int8")
	if err != nil {
		return 0, err
	}

	if num < math.MinInt8 || num > math.MaxInt8 {
		return 0, errValueOverflow(v, "int8")
	}

	return int8(num), nil
}

func (c primitiveConv) doPrimitiveToUint64(v interface{}, dstType string) (uint64, error) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
	switch {
	case kind == reflect.String:
		return strconv.ParseUint(val.String(), 0, 64)

	case kind == reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil

	case isKindInt(kind):
		num := val.Int()
		if num < 0 {
			return 0, errValueOverflow(v, dstType)
		}
		return uint64(num), nil

	case isKindUint(kind):
		return val.Uint(), nil

	case isKindFloat(kind):
		f := val.Float()
		return c.doFloatToUint(f, dstType)

	case isKindComplex(kind):
		// Prevent data loss, ensure the imaginary part is zero.
		cpl := val.Complex()
		partImag := imag(cpl)
		if partImag != 0 {
			return 0, errImaginaryPartLoss(v, dstType)
		}

		partReal := real(cpl)
		return c.doFloatToUint(partReal, dstType)
	}

	return 0, errCantConvertTo(v, dstType)
}

func (c primitiveConv) doFloatToUint(f float64, dstType string) (uint64, error) {
	if f < 0 || f > math.MaxUint64 {
		return 0, errValueOverflow(f, dstType)
	}

	if f != math.Trunc(f) {
		return 0, errPrecisionLoss(f, dstType)
	}

	return uint64(f), nil
}

func (c primitiveConv) toUint64(v interface{}) (uint64, error) {
	return c.doPrimitiveToUint64(v, "uint64")
}

func (c primitiveConv) toUint(v interface{}) (uint, error) {
	num, err := c.doPrimitiveToUint64(v, "uint")
	if err != nil {
		return 0, err
	}

	if num > maxUint {
		return 0, errValueOverflow(v, "uint")
	}

	return uint(num), nil
}

func (c primitiveConv) toUint32(v interface{}) (uint32, error) {
	num, err := c.doPrimitiveToUint64(v, "uint32")
	if err != nil {
		return 0, err
	}

	if num > math.MaxUint32 {
		return 0, errValueOverflow(v, "uint32")
	}

	return uint32(num), nil
}

func (c primitiveConv) toUint16(v interface{}) (uint16, error) {
	num, err := c.doPrimitiveToUint64(v, "uint16")
	if err != nil {
		return 0, err
	}

	if num > math.MaxUint16 {
		return 0, errValueOverflow(v, "uint16")
	}

	return uint16(num), nil
}

func (c primitiveConv) toUint8(v interface{}) (uint8, error) {
	num, err := c.doPrimitiveToUint64(v, "uint8")
	if err != nil {
		return 0, err
	}

	if num > math.MaxUint8 {
		return 0, errValueOverflow(v, "uint8")
	}

	return uint8(num), nil
}

func (c primitiveConv) doPrimitiveToFloat64(v interface{}, dstType string) (float64, error) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
	switch {
	case kind == reflect.String:
		return strconv.ParseFloat(val.String(), 64)

	case kind == reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil

	case isKindInt(kind):
		return float64(val.Int()), nil

	case isKindUint(kind):
		return float64(val.Uint()), nil

	case isKindFloat(kind):
		return val.Float(), nil

	case isKindComplex(kind):
		// Prevent data loss, ensure the imaginary part is zero.
		cpl := val.Complex()
		partImag := imag(cpl)
		if partImag != 0 {
			return 0, errImaginaryPartLoss(v, dstType)
		}
		return real(cpl), nil
	}

	return 0, errCantConvertTo(v, dstType)
}

func (c primitiveConv) toFloat64(v interface{}) (float64, error) {
	return c.doPrimitiveToFloat64(v, "float64")
}

func (c primitiveConv) toFloat32(v interface{}) (float32, error) {
	num, err := c.doPrimitiveToFloat64(v, "float32")
	if err != nil {
		return 0, err
	}

	if num < -math.MaxFloat32 || num > math.MaxFloat32 {
		return 0, errValueOverflow(v, "float32")
	}

	return float32(num), nil
}

func (c primitiveConv) doPrimitiveToComplex128(v interface{}, dstType string) (complex128, error) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
	switch {
	case kind == reflect.String:
		return strconv.ParseComplex(val.String(), 128)

	case kind == reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil

	case isKindInt(kind):
		return complex(float64(val.Int()), 0), nil

	case isKindUint(kind):
		return complex(float64(val.Uint()), 0), nil

	case isKindFloat(kind):
		return complex(val.Float(), 0), nil

	case isKindComplex(kind):
		return val.Complex(), nil
	}

	return 0, errCantConvertTo(v, dstType)
}

func (c primitiveConv) toComplex128(v interface{}) (complex128, error) {
	return c.doPrimitiveToComplex128(v, "complex128")
}

func (c primitiveConv) toComplex64(v interface{}) (complex64, error) {
	num, err := c.doPrimitiveToComplex128(v, "complex64")
	if err != nil {
		return 0, err
	}
	return complex64(num), nil
}
