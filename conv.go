// Package conv provides a group of functions to convert between primitive types, maps, slices and structs.
package conv

import (
	"fmt"
	"reflect"
	"time"
)

// Conv provides a group of functions to convert between simple types, maps, slices and structs.
// A pointer of a zero value is ready to use, it has the default behavior:
//
//	new(Conv).ConvertType(...)
//
// The field Config is used to customize the conversion behavior.
// e.g., this Conv instance uses a built-in FieldMatcherCreator and a custom TimeToString function.
//
//	c:= &Conv{
//	    Config: Config {
//	        FieldMatcherCreator: CaseInsensitiveFieldMatcherCreator(),
//	        TimeToString: func(t time.Time) (string, error) { return t.Format(time.RFC1123), nil },
//	    },
//	}
//
// All functions are thread-safe and can be used concurrently.
type Conv struct {
	// Conf is used to customize the conversion behavior.
	Conf Config
}

// Config is used to customize the conversion behavior of Conv .
type Config struct {
	// StringSplitter is the function used to split the string into elements of the slice when converting a string to a slice.
	// It is called internally by Convert(), ConvertType() or other functions.
	// Set this field if customization of the conversion is needed.
	// If this field is nil, the value will not be split.
	StringSplitter func(v string) []string

	// FieldMatcherCreator is used to get FieldMatcher instances when converting from map to struct or
	// from struct to struct.
	//
	// When converting a map to a struct, a FieldMatcherCreator.GetMatcher() returns a FieldMatcher instance for the
	// target struct, then FieldMatcher.MatchField() is applied to each key of the map.
	// The matched field will be set by the converted value of the key, the value is converted with Conv.ConvertType().
	//
	// When converting a struct to another, FieldMatcher.MatchField() is applied to each field name from the source struct.
	//
	// If FieldMatcherCreator is nil, SimpleMatcherCreator() will be used. There are some predefined implementations,
	// such as CaseInsensitiveFieldMatcherCreator(), CamelSnakeCaseFieldMatcherCreator().
	FieldMatcherCreator FieldMatcherCreator

	// CustomConverters provides a group of functions for converting the given value to some specific type.
	// The target type will never be nil.
	//
	// These functions are used to customize the conversion.
	// It is only used by Convert() or ConvertType(), not works in other functions.
	//
	// When a conversion starts, it will firstly go through each function in this slice:
	//   - The conversion stops immediately when some function returns a non-nil result or an error.
	//     Convert() or ConvertType() will use the result or returns the error directly.
	//   - The conversion runs next function in the slice if the previous one return nil with no error.
	//   - If no function in the slice returns OK, the conversion will continue with the predefined implementations,
	//     such as MapToMap(), StructToMap(), etc.
	//
	// NOTE: If your ConvertFunc use Conv internally, be carefully if there will be infinity loops. Is it suggested to
	// use a Conv instance with no ConvertFunc for the internal conversions.
	CustomConverters []ConvertFunc

	// TimeToString formats the given time.
	// It is called internally by Convert(), ConvertType() or other functions.
	// Set this field if it is needed to customize the procedure.
	// If this field is nil, the function DefaultTimeToString() will be used.
	TimeToString func(t time.Time) (string, error)

	// StringToTime parses the given string and returns the time it represents.
	// It is called internally by Convert, ConvertType or other functions.
	// Set this field if it is needed to customize the procedure.
	// If this field is nil, the function DefaultStringToTime() will be used.
	StringToTime func(v string) (time.Time, error)
}

// ConvertFunc is used to customize the conversion.
type ConvertFunc func(value interface{}, typ reflect.Type) (result interface{}, err error)

// DefaultTimeToString formats time using the time.RFC3339 format.
func DefaultTimeToString(t time.Time) (string, error) {
	return t.Format(time.RFC3339), nil
}

// DefaultStringToTime parses the time using the time.RFC3339Nano format.
func DefaultStringToTime(v string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, v)
}

func (c *Conv) doSplitString(v string) []string {
	var parts []string
	if c.Conf.StringSplitter == nil {
		parts = append(parts, v)
	} else {
		parts = c.Conf.StringSplitter(v)
	}
	return parts
}

func (c *Conv) doTimeToString(t time.Time) (string, error) {
	if c.Conf.TimeToString != nil {
		return c.Conf.TimeToString(t)
	}
	return DefaultTimeToString(t)
}

func (c *Conv) doStringToTime(v string) (time.Time, error) {
	if c.Conf.StringToTime != nil {
		return c.Conf.StringToTime(v)
	}
	return DefaultStringToTime(v)
}

// StringToSlice converts a string to a slice.
// The elements of the slice must be simple type, for which IsSimpleType() returns true.
//
// Conv.Config.StringSplitter() is used to split the string.
func (c *Conv) StringToSlice(v string, simpleSliceType reflect.Type) (interface{}, error) {
	const fnName = "StringToSlice"

	if simpleSliceType.Kind() != reflect.Slice {
		return nil, errForFunction(fnName, "the destination type must be slice, got %v", simpleSliceType)
	}

	elemTyp := simpleSliceType.Elem()
	if !IsSimpleType(elemTyp) {
		return nil, errForFunction(fnName, "cannot convert from string to %v, the element's type must be a simple type", simpleSliceType)
	}

	parts := c.doSplitString(v)
	dst := reflect.MakeSlice(simpleSliceType, 0, len(parts))
	for i, elemIn := range parts {
		elemOut, err := c.SimpleToSimple(elemIn, elemTyp)
		if err != nil {
			return nil, errForFunction(fnName, "cannot convert to %v, at index %v: %v", simpleSliceType, i, err)
		}

		dst = reflect.Append(dst, reflect.ValueOf(elemOut))
	}

	return dst.Interface(), nil
}

// SimpleToBool converts the value to bool.
// The value must be simple, for which IsSimpleType() returns true.
//
// Rules:
//   - nil: as false.
//   - Numbers: zero as false, non-zero as true.
//   - String: same as strconv.ParseBool().
//   - time.Time: zero Unix timestamps as false, other values as true.
//   - Other values are not supported, returns false and an error.
func (c *Conv) SimpleToBool(simple interface{}) (bool, error) {
	const fnName = "SimpleToBool"

	if simple == nil {
		return false, nil
	}

	typ := reflect.TypeOf(simple)
	if IsPrimitiveType(typ) {
		res, err := primitive.toBool(simple)
		if err == nil {
			return res, nil
		}
		return res, errForFunction(fnName, err.Error())
	}

	if typ == typTime {
		timestamp := simple.(time.Time).Unix()
		return timestamp != 0, nil
	}

	return false, errForFunction(fnName, "cannot convert %v to bool", typ)
}

// SimpleToString converts the given value to a string.
// The value must be a simple type, for which IsSimpleType() returns true.
//
// Conv.Config.StringToTime() is used to format times.
// Specially, booleans are converted to 0/1, not the default format true/false.
func (c *Conv) SimpleToString(v interface{}) (string, error) {
	const fnName = "SimpleToString"

	if v == nil {
		return "", errSourceShouldNotBeNil(fnName)
	}

	t := reflect.TypeOf(v)
	if t == typTime {
		res, err := c.doTimeToString(v.(time.Time))
		if err != nil {
			return "", errForFunction(fnName, "%s", err)
		}
		return res, nil
	}

	k := t.Kind()
	if !IsPrimitiveKind(k) {
		return "", errForFunction(fnName, "cannot convert %v to a primitive value", k)
	}

	return primitive.toString(v), nil
}

/*
SimpleToSimple converts a simple type, for which IsSimpleType() returns true, to another simple type.
The conversion use the following rules:

Booleans:
  - true/false is converted to number 0/1, or string '0'/'1'.
  - From a boolean to a string: use strconv.ParseBool().
  - From a number to a boolean: zero value as false; non-zero value as true.

Numbers:
  - From a complex number to a real number: the imaginary part must be zero, the real part will be converted.

To time.Time:
  - From a number: the number is treated as a Unix-timestamp as converted using time.Unix(),  the time zone is time.Local.
  - From a string: use Conv.Conf.StringToTime function.
  - From another time.Time: the raw value is cloned, includes the timestamp and the location.

From time.Time:
  - To a number: output a Unix-timestamp.
  - To a string: use Conv.Conf.TimeToString function.
*/
func (c *Conv) SimpleToSimple(src interface{}, dstTyp reflect.Type) (interface{}, error) {
	const fnName = "SimpleToSimple"

	if src == nil {
		return nil, errSourceShouldNotBeNil(fnName)
	}

	var res interface{}
	var err error
	dstKind := dstTyp.Kind()
	if IsPrimitiveKind(dstKind) {
		res, err = c.simpleToPrimitive(src, dstKind)
	} else if dstTyp.ConvertibleTo(typTime) {
		res, err = c.simpleToTime(src)
	} else {
		return nil, errForFunction(fnName, "cannot convert from %T to %v", src, dstTyp)
	}

	if err != nil {
		return nil, errForFunction(fnName, "%s", err)
	}

	// Convert if necessary.
	if reflect.TypeOf(res) != dstTyp {
		res = reflect.ValueOf(res).Convert(dstTyp).Interface()
	}
	return res, nil
}

/*
time.Time -> raw value
string -> Conv.Conf.StringToTime()
number as unix-timestamp -> Local time
*/
func (c *Conv) simpleToTime(src interface{}) (time.Time, error) {
	srcTyp := reflect.TypeOf(src)

	if srcTyp == typTime {
		return src.(time.Time), nil
	}

	switch {
	case srcTyp.Kind() == reflect.String:
		t, err := c.doStringToTime(src.(string))
		if err != nil {
			return zeroTime, err
		}
		return t, nil

	case IsPrimitiveType(srcTyp):
		timestamp, err := primitive.toPrimitive(src, reflect.Int64)
		if err != nil {
			return zeroTime, err
		}
		return time.Unix(timestamp.(int64), 0), nil // Get a local time.
	}

	// All simple types are processed in the switch block above, this line should never run.
	return zeroTime, errCantConvertTo(src, "time.Time")
}

func (c *Conv) simpleToPrimitive(src interface{}, dstKind reflect.Kind) (interface{}, error) {
	srcTyp := reflect.TypeOf(src)
	if IsPrimitiveType(srcTyp) {
		return primitive.toPrimitive(src, dstKind)
	}

	if srcTyp == typTime {
		tm := src.(time.Time)
		switch {
		case dstKind == reflect.String:
			return c.doTimeToString(tm)

		case IsPrimitiveKind(dstKind):
			timestamp := tm.Unix()
			return primitive.toPrimitive(timestamp, dstKind)
		}
	}

	return nil, fmt.Errorf("cannot convert from %v to %v", srcTyp, dstKind)
}

// SliceToSlice converts a slice to another slice.
//
// Each element will be converted using Conv.ConvertType() .
// A nil slice will be converted to a nil slice of the destination type.
// If the source value is nil interface{}, returns nil and an error.
func (c *Conv) SliceToSlice(src interface{}, dstSliceTyp reflect.Type) (interface{}, error) {
	const fnName = "SliceToSlice"

	if src == nil {
		return nil, errSourceShouldNotBeNil(fnName)
	}

	vSrcSlice := reflect.ValueOf(src)
	if vSrcSlice.Kind() != reflect.Slice {
		return nil, errForFunction(fnName, "src must be a slice, got %v", vSrcSlice.Kind())
	}

	if dstSliceTyp.Kind() != reflect.Slice {
		return nil, errForFunction(fnName, "the destination type must be slice, got %v", dstSliceTyp.Kind())
	}

	// A nil slice will be converted to a nil slice.
	if vSrcSlice.IsNil() {
		return reflect.Zero(dstSliceTyp).Interface(), nil
	}

	srcLen := vSrcSlice.Len()
	dstElemTyp := dstSliceTyp.Elem()
	vDstSlice := reflect.MakeSlice(dstSliceTyp, 0, srcLen)

	for i := 0; i < srcLen; i++ {
		vSrcElem := vSrcSlice.Index(i)
		srcElem := vSrcElem.Interface()
		vDstElem, err := c.ConvertType(srcElem, dstElemTyp)
		if err != nil {
			return nil, errForFunction(fnName, "cannot convert to %v, at index %v : %v", dstSliceTyp, i, err.Error())
		}

		vDstSlice = reflect.Append(vDstSlice, reflect.ValueOf(vDstElem))
	}

	return vDstSlice.Interface(), nil
}

// MapToStruct converts a map[string]interface{} to a struct.
//
// Each exported field of the struct is indexed using Conv.Config.FieldMatcherCreator().
func (c *Conv) MapToStruct(m map[string]interface{}, dstTyp reflect.Type) (interface{}, error) {
	const fnName = "MapToStruct"

	if m == nil {
		return nil, errSourceShouldNotBeNil(fnName)
	}

	k := dstTyp.Kind()
	if k != reflect.Struct {
		return nil, errForFunction(fnName, "the destination type must be struct, got %v", dstTyp)
	}

	dst := reflect.New(dstTyp).Elem()
	ctor := c.fieldMatcherCreator()
	mather := ctor.GetMatcher(dstTyp)

	for k, vm := range m {
		field, ok := mather.MatchField(k)
		if !ok {
			continue
		}

		fieldValue, err := getFieldValue(dst, field.Index)
		if err != nil {
			return nil, errForFunction(fnName, err.Error())
		}

		if !fieldValue.CanSet() {
			continue
		}

		vf, err := c.ConvertType(vm, field.Type)
		if err != nil {
			return nil, errForFunction(fnName, "error on converting field '%v': %v", field.Name, err.Error())
		}

		fieldValue.Set(reflect.ValueOf(vf))
	}

	return dst.Interface(), nil
}

func (c *Conv) fieldMatcherCreator() FieldMatcherCreator {
	g := c.Conf.FieldMatcherCreator
	if g == nil {
		g = new(SimpleMatcherCreator)
	}
	return g
}

// MapToMap converts a map to another map.
// If the source value is nil, the function returns a nil map of the destination type without any error.
//
// All keys and values in the map are converted using Conv.ConvertType() .
func (c *Conv) MapToMap(m interface{}, typ reflect.Type) (interface{}, error) {
	const fnName = "MapToMap"

	src := reflect.ValueOf(m)
	if src.Kind() != reflect.Map {
		return nil, errForFunction(fnName, "the given value type must be a map, got %v", src.Kind())
	}

	if typ.Kind() != reflect.Map {
		return nil, errForFunction(fnName, "the destination type must be map, got %v", typ)
	}

	if src.IsNil() {
		return reflect.Zero(typ).Interface(), nil
	}

	dst := reflect.MakeMap(typ)
	dstKeyType := typ.Key()
	dstValueType := typ.Elem()
	iter := src.MapRange()

	for iter.Next() {
		srcKey := iter.Key().Interface()
		dstKey, err := c.ConvertType(srcKey, dstKeyType)
		if err != nil {
			return nil, errForFunction(fnName, "cannot covert key '%v' to %v: %v", srcKey, dstKeyType, err.Error())
		}

		srcVal := iter.Value().Interface()
		dstVal, err := c.ConvertType(srcVal, dstValueType)
		if err != nil {
			return nil, errForFunction(fnName, "cannot covert value of key '%v' to %v: %v", srcKey, dstValueType, err.Error())
		}

		dst.SetMapIndex(reflect.ValueOf(dstKey), reflect.ValueOf(dstVal))
	}

	return dst.Interface(), nil
}

// StructToMap is partially like json.Unmarshal(json.Marshal(v), &someMap) . It converts a struct to map[string]interface{} .
//
// Each value of exported field will be processed recursively with an internal function f() , which:
//
// Simple types, for which IsSimpleType() returns true:
//   - A type whose kind is primitive, will be converted to a primitive value.
//   - For other types, the value will be cloned into the map directly.
//
// Slices:
//   - A nil slice is converted to a nil slice; an empty slice is converted to an empty slice with cap=0.
//   - A non-empty slice is converted to another slice, each element is process with f() , all elements must be the same type.
//
// Maps:
//   - A nil map are converted to nil of map[string]interface{} .
//   - A non-nil map is converted to map[string]interface{} , keys are processed with Conv.ConvertType() , values with f() .
//
// Structs are converted to map[string]interface{} using Conv.StructToMap() recursively.
//
// Pointers:
//   - Nils are ignored.
//   - Non-nil values pointed to are converted with f() .
//
// Other types not listed above are not supported and will result in an error.
func (c *Conv) StructToMap(v interface{}) (map[string]interface{}, error) {
	const fnName = "StructToMap"

	if v == nil {
		return nil, errSourceShouldNotBeNil(fnName)
	}

	srcTyp := reflect.TypeOf(v)
	if srcTyp.Kind() != reflect.Struct {
		return nil, errForFunction(fnName, "the given value must be a struct, got %v", srcTyp)
	}

	src := reflect.ValueOf(v)
	dst := reflect.MakeMap(reflect.TypeOf(map[string]interface{}(nil)))
	walker := NewFieldWalker(src.Type(), "") // TODO Tags on fields are not processed here.

	var err error
	walker.WalkValues(src, func(fi FieldInfo, fieldValue reflect.Value) bool {
		var ff reflect.Value
		ff, err = c.convertToMapValue(fieldValue)

		if err != nil {
			err = errForFunction(fnName, "error on converting field %v: %v", fi.Name, err.Error())
			return false
		}

		// If ff is nil value, the map index will not be set.
		dst.SetMapIndex(reflect.ValueOf(fi.Name), ff)
		return true
	})

	if err != nil {
		return nil, err
	}
	return dst.Interface().(map[string]interface{}), nil
}

func (c *Conv) convertToMapValue(fv reflect.Value) (reflect.Value, error) {
	for fv.Kind() == reflect.Ptr {
		fv = fv.Elem()
	}

	switch fv.Kind() {
	case reflect.Invalid:
		// Will be ignored in the outer loop.
		return reflect.ValueOf(nil), nil

	case reflect.Struct:
		v, err := c.StructToMap(fv.Interface())
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(v), nil

	case reflect.Slice:
		switch {
		case fv.IsNil():
			ft := fv.Type()
			sliceType, ok := c.determineSliceTypeForMapValue(ft)
			if !ok {
				return reflect.Value{}, fmt.Errorf("cannot convert %v", fv.Type())
			}
			return reflect.Zero(sliceType), nil

		case fv.Len() == 0:
			ft := fv.Type()
			sliceType, ok := c.determineSliceTypeForMapValue(ft)
			if !ok {
				return reflect.Value{}, fmt.Errorf("cannot convert %v", fv.Type())
			}
			return reflect.MakeSlice(sliceType, 0, 0), nil

		default:
			var newSlice reflect.Value

			for i := 0; i < fv.Len(); i++ {
				oldVal := fv.Index(i)
				newVal, err := c.convertToMapValue(oldVal)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("index %v: %v", i, err.Error())
				}

				// Lazy initialization. The slice type depends on the type of the first element.
				if i == 0 {
					newSlice = reflect.MakeSlice(reflect.SliceOf(newVal.Type()), 0, fv.Len())
				}

				newSlice = reflect.Append(newSlice, newVal)
			}

			return newSlice, nil
		}

	case reflect.Map:
		if fv.IsNil() {
			return reflect.ValueOf(map[string]interface{}(nil)), nil
		}

		newMap := reflect.MakeMap(reflect.TypeOf(map[string]interface{}(nil)))
		iter := fv.MapRange()
		for iter.Next() {
			oldKey := iter.Key()
			oldVal := iter.Value()

			var newKey string
			err := c.Convert(oldKey.Interface(), &newKey)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("key %v: %v", oldKey, err.Error())
			}

			newVal, err := c.convertToMapValue(oldVal)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("value of key %v: %v", newKey, err.Error())
			}

			newMap.SetMapIndex(reflect.ValueOf(newKey), newVal)
		}
		return newMap, nil

	case reflect.Interface:
		// Extract the underlying value.
		fv = reflect.ValueOf(fv.Interface())
		return c.convertToMapValue(fv)

	default:
		if IsPrimitiveKind(fv.Kind()) {
			res, err := c.simpleToPrimitive(fv.Interface(), fv.Kind())
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(res), nil
		}

		if !IsSimpleType(fv.Type()) {
			return reflect.Value{}, fmt.Errorf("must be a simple type, got %v", fv.Kind())
		}

		// Consider convert types which are simple but non-primitive - such as time.Time - to primitive types?
		return fv, nil
	}
}

func (c *Conv) determineSliceTypeForMapValue(srcSliceType reflect.Type) (dstSliceType reflect.Type, ok bool) {
	elemType := srcSliceType.Elem()
	if IsSimpleType(elemType) {
		dstSliceType = srcSliceType
		ok = true
		return
	}

	elemKind := elemType.Kind()
	switch elemKind {
	case reflect.Map, reflect.Struct:
		dstSliceType = reflect.SliceOf(reflect.TypeOf(map[string]interface{}(nil)))
		ok = true
		return

	case reflect.Slice:
		innerSliceType, innerOK := c.determineSliceTypeForMapValue(elemType)
		if !innerOK {
			return
		}

		dstSliceType = reflect.SliceOf(innerSliceType)
		ok = true
		return

	default:
		ok = false
		return
	}
}

// StructToStruct converts a struct to another.
// If the given value is nil, returns nil and an error.
//
// When converting, each field of the destination struct is indexed using Conv.Config.FieldMatcherCreator.
// The field values are converted using Conv.ConvertType() .
//
// This function can be used to deep-clone a struct.
func (c *Conv) StructToStruct(src interface{}, dstTyp reflect.Type) (interface{}, error) {
	const fnName = "StructToStruct"

	if src == nil {
		return nil, errSourceShouldNotBeNil(fnName)
	}

	dstKind := dstTyp.Kind()
	if dstKind != reflect.Struct {
		return nil, errForFunction(fnName, "the destination type must be struct, got %v", dstKind)
	}

	srcTyp := reflect.TypeOf(src)
	if srcTyp.Kind() != reflect.Struct {
		return nil, errForFunction(fnName, "the given value must be a struct, got %v", srcTyp)
	}

	ctor := c.fieldMatcherCreator()
	mather := ctor.GetMatcher(dstTyp)
	vSrc := reflect.ValueOf(src)
	vDst := reflect.New(dstTyp).Elem()
	walker := NewFieldWalker(vSrc.Type(), "") // TODO Tags on fields are not processed here.

	var err error
	walker.WalkValues(vSrc, func(fi FieldInfo, fieldValue reflect.Value) bool {
		field, ok := mather.MatchField(fi.Name)
		if !ok {
			return true
		}

		vField, e := getFieldValue(vDst, field.Index)
		if e != nil {
			err = errForFunction(fnName, e.Error())
			return false
		}

		if !vField.CanSet() {
			return true
		}

		dstValue, e := c.ConvertType(fieldValue.Interface(), vField.Type())
		if e != nil {
			err = errForFunction(fnName, "error on converting field %v: %v", field.Name, e.Error())
			return false
		}

		vField.Set(reflect.ValueOf(dstValue))
		return true
	})

	if err != nil {
		return nil, err
	}
	return vDst.Interface(), nil
}

// ConvertType is the core function of Conv . It converts the given value to the destination type.
//
// Currently, these conversions are supported:
//
//	simple                 -> simple                  use Conv.SimpleToSimple()
//	string                 -> []simple                use Conv.StringToSlice()
//	map[string]interface{} -> struct                  use Conv.MapToStruct()
//	map[ANY]ANY            -> map[ANY]ANY             use Conv.MapToMap()
//	[]ANY                  -> []ANY                   use Conv.SliceToSlice()
//	struct                 -> map[string]interface{}  use Conv.StructToMap()
//	struct                 -> struct                  use Conv.StructToStruct()
//
// 'ANY' generally can be any other type listed above. 'simple' is some type which IsSimpleType() returns true.
//
// If the destination type is the type of the empty interface, the function returns src directly without any error.
//
// For pointers:
// If the source value is a pointer, the value pointed to will be extracted and converted.
// The destination type can be a type of pointer, the source value which is nil will be converted to a nil pointer.
//
// This function can be used to deep-clone a struct, e.g.:
//
//	clone, err := ConvertType(src, reflect.TypeOf(src))
//
// There is a special conversion that can convert a map[string]interface{} to some other type listed above, when
// the map has only one key and the key is an empty string, the conversion is performed over the value other than
// the map itself. This is a special contract for some particular situation, when some code is working on maps only.
func (c *Conv) ConvertType(src interface{}, dstTyp reflect.Type) (interface{}, error) {
	const fnName = "ConvertType"

	if dstTyp == typEmptyInterface {
		return src, nil
	}

	// Convert nils to nil pointers.
	if src == nil && dstTyp.Kind() == reflect.Ptr {
		return reflect.Zero(dstTyp).Interface(), nil
	}

	// CustomConverters
	for i, f := range c.Conf.CustomConverters {
		res, err := f(src, dstTyp)
		if err != nil {
			return nil, errForFunction(fnName, "converter[%d]: %s", i, err.Error())
		}

		if res != nil {
			return res, nil
		}
	}

	// Try to get the underlying type from a pointer type.
	// It may be a pointer to another pointer, we should count the depth.
	ptrDepth := 0
	for dstTyp.Kind() == reflect.Ptr {
		dstTyp = dstTyp.Elem()
		ptrDepth++
	}

	dst, err := c.convertToNonPtr(src, dstTyp)
	if err != nil {
		return nil, errForFunction(fnName, err.Error())
	}

	// Convert to pointer if needed.
	if ptrDepth > 0 {
		var prev, current reflect.Value
		for i := 0; i < ptrDepth; i++ {
			if i == 0 {
				prev = reflect.ValueOf(dst)
			} else {
				prev = current
			}

			current = reflect.New(prev.Type())
			current.Elem().Set(prev)
		}

		dst = current.Interface()
	}

	return dst, nil
}

// Convert is like Conv.ConvertType() , but receives a pointer instead of a type.
// It stores the result in the value pointed to by dst.
//
// If the source value is nil, the function returns without an error, the underlying value
// of the pointer will not be set.
// If dst is not a pointer, the function panics an error.
func (c *Conv) Convert(src interface{}, dstPtr interface{}) error {
	const fnName = "Convert"

	dstValue := reflect.ValueOf(dstPtr)
	if dstValue.Kind() != reflect.Ptr {
		panic(errForFunction(fnName, "the destination value must be a pointer"))
	}

	if dstValue.IsZero() {
		panic(errForFunction(fnName, "the pointer must be initialized"))
	}

	if src == nil {
		return nil
	}

	// CustomConverters
	for i, f := range c.Conf.CustomConverters {
		if dstValue.Kind() == reflect.Ptr {
			dstValue = dstValue.Elem()
		}

		res, err := f(src, dstValue.Type())
		if err != nil {
			return errForFunction(fnName, "converter[%d]: %s", i, err.Error())
		}

		if res != nil {
			dstValue.Set(reflect.ValueOf(res))
			return nil
		}
	}

	for dstValue.Kind() == reflect.Ptr {
		dstValue = dstValue.Elem()
		if dstValue.Kind() == reflect.Invalid {
			panic(errForFunction(fnName, "the underlying pointer must be initialized"))
		}
	}

	dstTyp := dstValue.Type()
	value, err := c.convertToNonPtr(src, dstTyp)
	if err != nil {
		return errForFunction(fnName, err.Error())
	}

	dstValue.Set(reflect.ValueOf(value))
	return nil
}

// MustConvertType is like ConvertType() but panics instead of returns an error.
func (c *Conv) MustConvertType(src interface{}, dstTyp reflect.Type) interface{} {
	res, err := c.ConvertType(src, dstTyp)
	if err != nil {
		panic(err)
	}
	return res
}

// MustConvert is like Convert() but panics instead of returns an error.
func (c *Conv) MustConvert(src interface{}, dstPtr interface{}) {
	err := c.Convert(src, dstPtr)
	if err != nil {
		panic(err)
	}
}

// getUnderlyingValue extracts the underlying value if v is a pointer; otherwise returns v.
// If the pointer points to nil, returns nil.
func (c *Conv) getUnderlyingValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	vo := reflect.ValueOf(v)
	for vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}

	// Nil for nil pointer.
	if vo.Kind() == reflect.Invalid {
		return nil
	}

	return vo.Interface()
}

func (c *Conv) convertToNonPtr(src interface{}, dstTyp reflect.Type) (interface{}, error) {
	src = c.getUnderlyingValue(src)

	dstKind := dstTyp.Kind()
	if src == nil {
		if dstKind == reflect.Slice || dstKind == reflect.Map {
			return reflect.Zero(dstTyp).Interface(), nil
		}
		return nil, fmt.Errorf("cannot convert nil to %v", dstTyp)
	}

	srcTyp := reflect.TypeOf(src)
	srcKind := srcTyp.Kind()
	if IsSimpleType(srcTyp) && IsSimpleType(dstTyp) {
		return c.SimpleToSimple(src, dstTyp)
	}

	if srcKind == reflect.Map {
		// map[string]ANY { "": value } -> ConvertType(value)
		if underlyingValue := c.tryFlattenEmptyKeyMap(src); underlyingValue != nil {
			return c.ConvertType(underlyingValue, dstTyp)
		}

		switch dstKind {
		// map -> map
		case reflect.Map:
			return c.MapToMap(src, dstTyp)

		// map[string]ANY -> struct
		case reflect.Struct:
			mm, ok := src.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("when converting a map to a struct, the map must be map[string]interface{}, got %v", srcTyp)
			}
			return c.MapToStruct(mm, dstTyp)
		}
	} else if srcKind == reflect.Struct {
		switch dstKind {
		case reflect.Map:
			if dstTyp != typStringMap {
				return nil, fmt.Errorf("when converting a struct to a map, the destination type must be map[string]interface{}, got %v", dstTyp)
			}
			return c.StructToMap(src)

		case reflect.Struct:
			return c.StructToStruct(src, dstTyp)
		}
	} else if dstKind == reflect.Slice {
		switch srcKind {
		// string -> []simple
		case reflect.String:
			return c.StringToSlice(src.(string), dstTyp)

		case reflect.Slice:
			return c.SliceToSlice(src, dstTyp)
		}
	}

	return nil, fmt.Errorf("cannot convert %v to %v", srcTyp, dstTyp)
}

// tryFlattenEmptyKeyMap check the value. When all those conditions are satisfied:
//   - the map is map[string]interface{}
//   - the map has only one key
//   - the key is an empty string
//
// The function returns the value of the key; otherwise it returns nil.
//
// Such map is a special contract, it's used when converting a map to a simple type.
// e.g., map[string]int{"": 123} can be converted to 123 .
func (c *Conv) tryFlattenEmptyKeyMap(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok || len(m) != 1 {
		return nil
	}

	for k, v := range m {
		if k == "" {
			return v
		}
	}

	return nil
}
