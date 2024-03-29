package conv_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cmstar/go-conv"
)

type DemoUser struct {
	Name     string
	MailAddr string
	Age      int
	IsVip    bool
}

func Example() {

	// There is a group of shortcut functions for converting from/to simple types.
	// All conversion functions returns the converted value and an error.
	fmt.Println(conv.Int("123"))         // -> 123
	fmt.Println(conv.String(3.14))       // -> "3.14"
	fmt.Println(conv.Float64("invalid")) // -> get an error

	// When converting integers, overflow-checking is applied.
	fmt.Println(conv.Int8(1000)) // -> overflow

	// A zero value of number is converted to false; a non-zero value is converted to true.
	fmt.Println(conv.Bool(float32(0.0))) // -> false
	fmt.Println(conv.Bool(float64(0.1))) // -> true
	fmt.Println(conv.Bool(-1))           // -> true

	// strconv.ParseBool() is used for string-to-bool conversion.
	fmt.Println(conv.Bool("true"))  // -> true
	fmt.Println(conv.Bool("false")) // -> false

	// Numbers can be converted time.Time, they're treated as UNIX timestamp.
	// NOTE: The location of the time depends on the input type, see the example of time for more details.
	t, err := conv.Time(3600) // An hour later after 1970-01-01T00:00:00Z.
	// By default time.Time is converted to string with the RFC3339 format.
	fmt.Println(conv.String(t.UTC())) // -> 1970-01-01T01:00:00Z

	// ConvertType() is the core function in the package. In fact all conversion can be done via
	// this function. For complex types, there is no shortcut, we can use ConvertType() directly.
	// It receives the source value and the destination type.

	// Convert from a slice to another. ConvertType() is applied to each element.
	sliceOfInt, err := conv.ConvertType([]string{"1", "2", "3"}, reflect.TypeOf([]int{}))
	fmt.Println(sliceOfInt, err) // -> []int{1, 2, 3}

	// Convert from a map[string]interface{} to a struct. ConvertType() is applied to each field.
	user := DemoUser{Name: "Bob", MailAddr: "bob@example.org", Age: 51}
	out, err := conv.ConvertType(user, reflect.TypeOf(map[string]interface{}{}))
	fmt.Println(out, err) // -> map[string]interface{}{"Age":51, "MailAddr":"bob@example.org", "Name":"Bob", "IsVip":false}

	// From map to struct.
	m := map[string]interface{}{"Name": "Alice", "Age": "27", "IsVip": 1}
	out, err = conv.ConvertType(m, reflect.TypeOf(user))
	fmt.Printf("%+v\n", out) // -> DemoUser{Name: "Alice", MailAddr: "", Age: 27, IsVip:true}

	// Deep-clone a struct.
	out, err = conv.ConvertType(user, reflect.TypeOf(user))
	fmt.Printf("%+v\n", out) // -> DemoUser{Name: "Bob", MailAddr: "bob@example.org", Age: 51}

	// Convert() is similar to ConvertType(), but receive a pointer instead of a type.
	// It's more like some functions in the standard library such as json.Unmarshal().
	clone := DemoUser{}
	conv.Convert(user, &clone)
	fmt.Printf("%+v\n", clone) // -> DemoUser{Name: "Bob", MailAddr: "bob@example.org", Age: 51}

	// Output:
	// 123 <nil>
	// 3.14 <nil>
	// 0 strconv.ParseFloat: parsing "invalid": invalid syntax
	// 0 value overflow when converting 1000 (int) to int8
	// false <nil>
	// true <nil>
	// true <nil>
	// true <nil>
	// false <nil>
	// 1970-01-01T01:00:00Z <nil>
	// [1 2 3] <nil>
	// map[Age:51 IsVip:false MailAddr:bob@example.org Name:Bob] <nil>
	// {Name:Alice MailAddr: Age:27 IsVip:true}
	// {Name:Bob MailAddr:bob@example.org Age:51 IsVip:false}
	// {Name:Bob MailAddr:bob@example.org Age:51 IsVip:false}
}

func Example_timeConversion() {
	// When converting to a time.Time, the location depends on the input value.

	// When converting a number to a Time, returns a Time with the location time.Local, behaves like time.Unix().
	t, _ := conv.Time(0) // Shortcut of conv.ConvertType(0, reflect.TypeOf(time.Time{})) .

	fmt.Println("convert a number to a Time, the output location is:", t.Location()) // -> Local

	// Returns a Time with offset 02:00, like time.Parse().
	t, _ = conv.Time("2022-04-20T11:30:40+02:00")
	_, offset := t.Zone()
	fmt.Println(offset) // -> 7200 which is the total seconds of 02:00.

	// Returns  a Time with time.UTC, like time.Parse().
	t, _ = conv.Time("2022-04-20T11:30:40Z")
	fmt.Println(t.Location()) // -> UTC

	// When converting a Time to a Time, keeps the location: the raw value is cloned.
	t, _ = conv.Time(time.Now())
	fmt.Println("convert a Local time to a Local time, the output location is:", t.Location()) // -> Local
	t, _ = conv.Time(time.Now().UTC())
	fmt.Println("convert an UTC time to an UTC time, the output location is:", t.Location()) // -> UTC

	// When converting a Time to a number, outputs an UNIX timestamp.
	t, _ = time.Parse(time.RFC3339, "2022-04-20T03:30:40Z")
	fmt.Println(conv.MustInt(t)) // -> 1650425440

	// When converting a Time to a string, outputs using conv.DefaultTimeToString function, which format the time in RFC3339.
	fmt.Println(conv.MustString(t)) // -> 2022-04-20T03:30:40Z

	// Output:
	// convert a number to a Time, the output location is: Local
	// 7200
	// UTC
	// convert a Local time to a Local time, the output location is: Local
	// convert an UTC time to an UTC time, the output location is: UTC
	// 1650425440
	// 2022-04-20T03:30:40Z
}

func Example_theConvInstance() {

	// The Conv struct providers the underlying features for all shortcuts functions.
	// You can use it directly. A zero value has the default behavior.
	c := new(conv.Conv)

	// It has a field named Conf which is used to customize the conversion.
	// By default, we get an error when converting to a slice from a string that is a group of
	// elements separated by some characters.
	fmt.Println(c.ConvertType("1,2,3", reflect.TypeOf([]int{}))) // -> error

	// Conf.StringSplitter is a function that defines how to convert from a string to a slice.
	c.Conf.StringSplitter = func(v string) []string { return strings.Split(v, ",") }
	// After configure, the conversion should be OK.
	fmt.Println(c.ConvertType("1,2,3", reflect.TypeOf([]int{}))) // -> []int{1, 2, 3}

	// Conf.FieldMatcherCreator define how to match names from a struct when converting from
	// a map or another struct.
	// Here we demonstrate how to make snake-case names match the field names automatically,
	// using the build-in FieldMatcherCreator named SimpleMatcherCreator.
	c.Conf.FieldMatcherCreator = &conv.SimpleMatcherCreator{
		Conf: conv.SimpleMatcherConfig{
			CamelSnakeCase: true,
		},
	}
	// When then CamelSnakeCase option is true, 'mailAddr' can match the field MailAddr, 'is_vip' can match IsVip.
	m := map[string]interface{}{"name": "Bob", "age": "51", "mailAddr": "bob@example.org", "is_vip": "true"}
	user := DemoUser{}
	_ = c.Convert(m, &user)
	fmt.Printf("%+v\n", user) // -> DemoUser{Name: "Bob", MailAddr: "bob@example.org", Age: 51, IsVip: true})

	// The json package of the standard library does not support matching fields in case-insensitive manner.
	// We have to use field tag to specify the name of JSON properties.
	// With FieldMatcherCreator, we can unmarshal JSON in case-insensitive manner, using a map as a middleware.
	// Thou the performance is not good, but it works :) .
	middleware := make(map[string]interface{})
	_ = json.Unmarshal([]byte(`{"name":"Alice", "mailAddr":"alice@example.org", "isVip": true, "age":27}`), &middleware)
	_ = c.Convert(middleware, &user)
	fmt.Printf("%+v\n", user) // -> DemoUser{Name: "Alice", MailAddr: "alice@example.org", Age: 27, IsVip: true})

	// Output:
	// <nil> conv.ConvertType: conv.StringToSlice: cannot convert to []int, at index 0: conv.SimpleToSimple: strconv.ParseInt: parsing "1,2,3": invalid syntax
	// [1 2 3] <nil>
	// {Name:Bob MailAddr:bob@example.org Age:51 IsVip:true}
	// {Name:Alice MailAddr:alice@example.org Age:27 IsVip:true}
}

func Example_customConverters() {
	// We can use conv.ConvertFunc to convert values to some type of your own.

	type Name struct{ FirstName, LastName string }

	// Converts a string such as "A B" to a struct Name{A, B}.
	nameConverter := func(value interface{}, typ reflect.Type) (result interface{}, err error) {
		if typ != reflect.TypeOf(Name{}) {
			return nil, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, nil
		}

		parts := strings.Split(s, " ")
		if len(parts) != 2 {
			return nil, errors.New("bad name")
		}

		return Name{parts[0], parts[1]}, nil
	}

	c := &conv.Conv{
		Conf: conv.Config{
			CustomConverters: []conv.ConvertFunc{nameConverter},
		},
	}

	var name Name
	c.Convert("John Doe", &name)
	fmt.Println("FirstName:", name.FirstName)
	fmt.Println("LastName:", name.LastName)

	// The middle name is not supported so we'll get an error here.
	fmt.Println("error:")
	_, err := c.ConvertType("John Middle Doe", reflect.TypeOf(Name{}))
	fmt.Println(err)

	// Output:
	// FirstName: John
	// LastName: Doe
	// error:
	// conv.ConvertType: converter[0]: bad name
}
