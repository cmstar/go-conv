# conv - Go 的类型转换库

[![GoDoc](https://godoc.org/github.com/cmstar/go-conv?status.svg)](https://pkg.go.dev/github.com/cmstar/go-conv)
[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)

功能:

- 支持数值类型、日期、字符串间的转换。
- 支持 `slice` 到 `slice` ， `struct` 和 `map[string]interface{}` 间的转换。
- 整数转换时的溢出检验。
- `map` 到 `struct` 的转换中，可以指定以大小写不敏感的形式匹配名称，或以驼峰（camel-case）风格自动匹配名称。
- 利用一个 `map` 做中间件，可以实现字段大小写不敏感的 JSON 反序列化。
- 深拷贝（deep-clone）。
- 支持指针与非指针间的转换。
- 没有第三方依赖，仅使用标准库。

仍有功能在开发中，部分 API 可能在未来发生变化。

## 快速使用

安装：
```
go get -u github.com/cmstar/go-conv@latest
```

开始使用，下面的代码来自 example_test.go ：
```go
// There is a group of shortcut functions for converting from/to simple types.
// All conversion functions returns the converted value and an error.
fmt.Println(conv.Int("123"))         // -> 123
fmt.Println(conv.String(3.14))       // -> "3.14"
fmt.Println(conv.Float64("invalid")) // -> get an error

// When converting intergers, overflow-checking is applied.
fmt.Println(conv.Int8(1000)) // -> overflow

// A zero value of number is converted to false; a non-zero value is converted to true.
fmt.Println(conv.Bool(float32(0.0))) // -> false
fmt.Println(conv.Bool(float64(0.1))) // -> true
fmt.Println(conv.Bool(-1))           // -> true

// strconv.ParseBool() is used for string-to-bool conversion.
fmt.Println(conv.Bool("true"))  // -> true
fmt.Println(conv.Bool("false")) // -> false

// Numbers can be converted time.Time, they're treated as UNIX timestamp.
// NOTE: The location of the time output is **local**.
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
// Is't more like some functions in the standard library such as json.Unmarshal().
clone := DemoUser{}
conv.Convert(user, &clone)
fmt.Printf("%+v\n", clone) // -> DemoUser{Name: "Bob", MailAddr: "bob@example.org", Age: 51}

```

Output:
```
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
```

深入 Conv 实例:
```go
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
c.Conf.FieldMatcherCreator = conv.SimpleMatcherCreator{
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
// Thou the performace is not good, but it works :) .
middleware := make(map[string]interface{})
_ = json.Unmarshal([]byte(`{"name":"Alice", "mailAddr":"alice@example.org", "isVip": true, "age":27}`), &middleware)
_ = c.Convert(middleware, &user)
fmt.Printf("%+v\n", user) // -> DemoUser{Name: "Alice", MailAddr: "alice@example.org", Age: 27, IsVip: true})
```

Output:
```
// <nil> conv.ConvertType: conv.StringToSlice: cannot convert to []int, at index 0: conv.SimpleToSimple: strconv.ParseInt: parsing "1,2,3": invalid syntax
// [1 2 3] <nil>
// {Name:Bob MailAddr:bob@example.org Age:51 IsVip:true}
// {Name:Alice MailAddr:alice@example.org Age:27 IsVip:true}
```

## 性能

由于代码大量使用反射，不适合对于性能很敏感的场景。

## 已知问题

- 不能很好的支持内嵌结构。目前只被当做一个普通字段。
- 从 `struct` 转到 `map` 时，不支持字段上的 tag 。