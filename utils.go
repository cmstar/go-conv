package conv

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// The alias of the empty interface. Go 1.18 defines this but in earlier versions we can't use it.
type any = interface{}

var (
	minInt   int64
	maxInt   int64
	maxUint  uint64
	typTime  = reflect.TypeOf(time.Time{})
	zeroTime = time.Time{}

	// The type of map used when converting between structs and maps.
	typStringMap = reflect.TypeOf(map[string]interface{}(nil))
)

func init() {
	switch strconv.IntSize {
	case 32:
		minInt = math.MinInt32
		maxInt = math.MaxInt32
		maxUint = math.MaxUint32

	case 64:
		minInt = math.MinInt64
		maxInt = math.MaxInt64
		maxUint = math.MaxUint64
	}
}

// IsPrimitiveKind returns true if the given Kind is any of bool, int*, uint*, float*, complex* or string.
func IsPrimitiveKind(k reflect.Kind) bool {
	// Exclude reflect.Uintptr .
	return k >= reflect.Bool && k <= reflect.Uint64 ||
		k >= reflect.Float32 && k <= reflect.Complex128 ||
		k == reflect.String
}

// IsPrimitiveType returns true if the given type is any of bool, int*, uint*, float*, complex* or string.
func IsPrimitiveType(t reflect.Type) bool {
	return t != nil && IsPrimitiveKind(t.Kind())
}

// IsSimpleType returns true if the given type IsPrimitiveType() or is convertible to time.Time .
func IsSimpleType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	return IsPrimitiveType(t) || t.ConvertibleTo(typTime)
}

func isKindInt(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func isKindUint(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uint64
}

func isKindFloat(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

func isKindComplex(k reflect.Kind) bool {
	return k == reflect.Complex64 || k == reflect.Complex128
}

func errCantConvertTo(v interface{}, dstType string) error {
	return fmt.Errorf("cannot convert %#v (%[1]T) to %s", v, dstType)
}

func errValueOverflow(v interface{}, dstType string) error {
	return fmt.Errorf("value overflow when converting %#v (%[1]T) to %s", v, dstType)
}

func errPrecisionLoss(v interface{}, dstType string) error {
	return fmt.Errorf("lost precision when converting %#v (%[1]T) to %s", v, dstType)
}

func errImaginaryPartLoss(v interface{}, dstType string) error {
	return fmt.Errorf("lost imaginary part when converting %#v (%[1]T) to %s", v, dstType)
}

// errForFunction returns an error which is used by exported functions,
// the error message contains the function name.
func errForFunction(fn, msgFormat string, a ...interface{}) error {
	msg := "conv." + fn + ": " + fmt.Sprintf(msgFormat, a...)
	return errors.New(msg)
}

func errSourceShouldNotBeNil(fnName string) error {
	return errForFunction(fnName, "the source value should not be nil")
}

// getFieldPath returns the path of an embedded field. Embedded pointers are supported.
// Panics on invalid parameters.
//
// e.g.
//
//	type A struct { X, Y int }
//	type B struct { *A }
//	type C struct { B }
//	type D struct { *C }
//
//	getFieldPath(reflect.TypeOf(D{}), []int{0, 0, 0, 1}) //-> C.B.A.Y
func getFieldPath(typ reflect.Type, index []int) string {
	var path string
	for i := 0; i < len(index); i++ {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		fs := typ.Field(index[i])
		typ = fs.Type

		if i > 0 {
			path += "."
		}
		path += fs.Name
	}
	return path
}

// getFieldValue returns the value of the field at the given index.
// If the field is a field of an embedded pointer of a struct, and the pointer is nil, this function will try to
// initialize the value of the embedded pointer with the zero value.
//
// If the embedded pointer cannot be initialized, returns an error.
//
// If the value is nil, or is not an *addressable* struct, or the index is out of range,
// or the length of index is 0, the function panics.
func getFieldValue(val reflect.Value, index []int) (reflect.Value, error) {
	ln := len(index)
	if ln == 0 {
		panic("index must be given")
	}

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			panic("value is nil")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic("value must be struct")
	}

	current := val

	// It the field is embedded, ensure the parent struct has value.
	for i := 0; i < ln-1; i++ {
		current = current.Field(index[i])

		// The field may be a nested pointer such as **struct{...}, the check should be performed recursively.
		for {
			if current.Kind() != reflect.Ptr {
				goto next
			}

			if current.IsNil() {
				if !current.CanSet() {
					return reflect.Value{}, fmt.Errorf("cannot set embedded pointer on field %s", getFieldPath(val.Type(), index[:i+1]))
				}

				// Initialize with the zero value of the underlying type.
				v := reflect.New(current.Type().Elem())
				current.Set(v)
			}

			current = current.Elem()
		}
	next:
	}

	return current.Field(index[ln-1]), nil
}
