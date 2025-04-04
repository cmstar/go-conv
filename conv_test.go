package conv

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

type FromString string
type FromInt int

var _caseInsensitiveConv = &Conv{
	Conf: Config{
		FieldMatcherCreator: &SimpleMatcherCreator{
			Conf: SimpleMatcherConfig{
				CaseInsensitive: true,
			},
		},
	},
}

var _tagConv = &Conv{
	Conf: Config{
		FieldMatcherCreator: &SimpleMatcherCreator{
			Conf: SimpleMatcherConfig{
				Tag: "conv",
			},
		},
	},
}

func TestConv_StringToSlice(t *testing.T) {
	customConv := &Conv{
		Conf: Config{
			StringSplitter: func(v string) []string { return strings.Split(v, "~") },
		},
	}

	type args struct {
		v               string
		simpleSliceType reflect.Type
	}
	tests := []struct {
		name          string
		useCustomConv bool
		args          args
		want          interface{}
		errRegex      string
	}{
		// Test default splitting.
		{"string", false, args{"1~2~3", reflect.TypeOf([]string{})}, []string{"1~2~3"}, ""},

		// Test custom split function.
		{"string-empty", true, args{"", reflect.TypeOf([]string{})}, []string{""}, ""},
		{"string", true, args{"a", reflect.TypeOf([]string{})}, []string{"a"}, ""},
		{"string-slice", true, args{"a~b~c", reflect.TypeOf([]string{})}, []string{"a", "b", "c"}, ""},
		{"float-slice", true, args{"1.1~2.2~3.3", reflect.TypeOf([]float32{})}, []float32{1.1, 2.2, 3.3}, ""},
		{"bool", true, args{"1~true~False", reflect.TypeOf([]bool{})}, []bool{true, true, false}, ""},
		{"err", true, args{"1~x~0", reflect.TypeOf([]bool{})}, nil, "^conv.StringToSlice: .+, at index 1: .+"},
		{"float-slice", true, args{"1.1~2.2~3.3", reflect.TypeOf([]float32{})}, []float32{1.1, 2.2, 3.3}, ""},

		{"err-src", false, args{"", reflect.TypeOf(1)}, nil, "must be slice"},
		{"err-elem", false, args{"", reflect.TypeOf([]struct{}{})}, nil, "must be a simple type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustomConv {
				got, err = customConv.StringToSlice(tt.args.v, tt.args.simpleSliceType)
			} else {
				got, err = _defaultConv.StringToSlice(tt.args.v, tt.args.simpleSliceType)
			}

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("StringToSlice() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("StringToSlice() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_SimpleToBool(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"nil", args{nil}, false, false},
		{"true", args{true}, true, false},
		{"false", args{false}, false, false},
		{"0", args{0}, false, false},
		{"1", args{1}, true, false},
		{"-100", args{-100}, true, false},
		{"3000", args{uint64(3000)}, true, false},
		{"6553.6", args{6553.6}, true, false},
		{"0.0", args{0.0}, false, false},
		{"0+0i", args{0 + 0i}, false, false},
		{"time-false", args{time.Now()}, true, false},
		{"time-true", args{time.Unix(0, 0)}, false, false},
		{"string-true", args{"true"}, true, false},
		{"string-False", args{"False"}, false, false},
		{"string-0", args{"0"}, false, false},
		{"string-1", args{"1"}, true, false},
		{"err-string", args{"wrong"}, false, true},
		{"err-struct", args{struct{}{}}, false, true},
		{"err-slice", args{[]int{}}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _defaultConv.SimpleToBool(tt.args.v)
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

func TestConv_SimpleToString(t *testing.T) {
	customTimeConv := &Conv{
		Conf: Config{
			TimeToString: func(t time.Time) (string, error) {
				if t == time.Unix(0, 0) {
					return "", errors.New("we make a custom error for zero time")
				}
				return t.Format("1-02,2006 15:04:05.000"), nil
			},
		},
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name        string
		useCustConv bool
		args        args
		want        string
		wantErr     bool
	}{
		{"true", false, args{true}, "1", false},
		{"false", false, args{false}, "0", false},
		{"int", false, args{-112334}, "-112334", false},

		{"float32", false, args{float32(3.14)}, "3.14", false},
		{"float32-large", false, args{float32(11122233)}, "11122233", false},
		{"float64-large", false, args{float64(111222333444.555)}, "111222333444.555", false},

		{"string", false, args{"中"}, "中", false},
		{"time0", false, args{time.Date(2020, 1, 20, 13, 6, 22, int(321*time.Millisecond), time.UTC)}, "2020-01-20T13:06:22Z", false},
		{"time+5", false, args{time.Date(2020, 1, 20, 13, 6, 22, int(321*time.Millisecond), time.FixedZone("", 5*3600))}, "2020-01-20T13:06:22+05:00", false},

		{"nil", false, args{nil}, "", true},
		{"err1", false, args{struct{}{}}, "", true},
		{"err2", false, args{map[string]interface{}{}}, "", true},
		{"err-cust", true, args{time.Unix(0, 0)}, "", true},

		// Customized Conv.Config.TimeToString() .
		{"cus-time", true, args{time.Date(2020, 1, 6, 13, 6, 22, int(321*time.Millisecond), time.UTC)}, "1-06,2020 13:06:22.321", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = customTimeConv.SimpleToString(tt.args.v)
			} else {
				got, err = _defaultConv.SimpleToString(tt.args.v)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_SimpleToSimple(t *testing.T) {
	spUtcTime := time.Date(2021, 6, 3, 13, 21, 22, 54321, time.UTC)
	spUtcTimeWithoutNano := time.Unix(spUtcTime.Unix(), 0).UTC()
	spLocalTime := spUtcTime.Local()

	customTimeConv := &Conv{
		Conf: Config{
			StringToTime: func(v string) (time.Time, error) { return spLocalTime, nil },
			TimeToString: func(t time.Time) (string, error) {
				if t == time.Unix(0, 0) {
					return "", errors.New("we make a custom error for zero time")
				}
				return t.Format("20060102"), nil
			},
		},
	}

	type Empty struct{}

	type args struct {
		src     interface{}
		dstType reflect.Type
	}
	tests := []struct {
		name        string
		useCustConv bool
		args        args
		want        interface{}
		errRegex    string
	}{
		// primitive to primitive
		{"int-int", false, args{1, reflect.TypeOf(0)}, 1, ""},
		{"int-int64", false, args{1, reflect.TypeOf(int64(1))}, int64(1), ""},
		{"float-float", false, args{3.3, reflect.TypeOf(3.3)}, 3.3, ""},
		{"float32-uint32", false, args{float32(123456), reflect.TypeOf(uint32(0))}, uint32(123456), ""},
		{"int-string", false, args{-25, reflect.TypeOf("")}, "-25", ""},
		{"string-string", false, args{"some", reflect.TypeOf("")}, "some", ""},
		{"string-int", false, args{"789", reflect.TypeOf(0)}, 789, ""},
		{"string-true", false, args{"true", reflect.TypeOf(true)}, true, ""},
		{"string-false", false, args{"0", reflect.TypeOf(true)}, false, ""},
		{"false-int", false, args{false, reflect.TypeOf(0)}, 0, ""},
		{"true-true", false, args{true, reflect.TypeOf(true)}, true, ""},
		{"true-float", false, args{true, reflect.TypeOf(1.0)}, 1.0, ""},
		{"true-string", false, args{true, reflect.TypeOf("")}, "1", ""},
		{"false-string", false, args{false, reflect.TypeOf("")}, "0", ""},
		{"complex128-complex128", false, args{complex128(12.5 + 33i), reflect.TypeOf(complex128(0 + 0i))}, complex128(12.5 + 33i), ""},
		{"complex64-complex64", false, args{complex64(12.5 + 33i), reflect.TypeOf(complex64(0 + 0i))}, complex64(12.5 + 33i), ""},
		{"complex64-complex128", false, args{complex64(12.5 + 33i), reflect.TypeOf(complex128(0 + 0i))}, complex128(12.5 + 33i), ""},
		{"true-complex128", false, args{true, reflect.TypeOf(complex128(0 + 0i))}, complex128(1 + 0i), ""},
		{"false-complex64", false, args{false, reflect.TypeOf(complex64(0 + 0i))}, complex64(0 + 0i), ""},
		{"complex128-true", false, args{complex128(-5 + 3i), reflect.TypeOf(false)}, true, ""},
		{"complex64-float64", false, args{complex64(12.5 + 0i), reflect.TypeOf(0.0)}, 12.5, ""},
		{"complex128-uint16", false, args{complex128(5544 + 0i), reflect.TypeOf(uint16(0))}, uint16(5544), ""},

		// time
		{"utc-local", false, args{spUtcTime, reflect.TypeOf(time.Time{})}, spUtcTime, ""},
		{"time-string", false, args{spLocalTime, reflect.TypeOf("")}, spLocalTime.Format(time.RFC3339), ""},
		{"time-string-custom", true, args{spLocalTime, reflect.TypeOf("")}, "20210603", ""},
		{"string-time", false, args{"2021-06-03T13:21:22Z", reflect.TypeOf(time.Time{})}, spUtcTimeWithoutNano, ""},
		{"string-time-custom", true, args{"any", reflect.TypeOf(time.Time{})}, spLocalTime, ""}, // always returns spLocalTime
		{"time-int", false, args{spLocalTime, reflect.TypeOf(0)}, int(spLocalTime.Unix()), ""},
		{"time-float", false, args{spLocalTime, reflect.TypeOf(0.0)}, float64(spLocalTime.Unix()), ""},
		{"int-time", false, args{1622726482, reflect.TypeOf(time.Time{})}, spUtcTimeWithoutNano.Local(), ""},

		// err
		{"err-nil", false, args{nil, reflect.TypeOf(1)}, nil, "^conv.SimpleToSimple: the source value should not be nil$"},
		{"err-time-from-string", false, args{"date", reflect.TypeOf(time.Time{})}, nil, "^conv.SimpleToSimple: .+"},
		{"err-time-from-complex", false, args{1 + 3i, reflect.TypeOf(time.Time{})}, nil, "lost imaginary part"},
		{"err-time-to-int8", false, args{spLocalTime, reflect.TypeOf(int8(0))}, nil, `value overflow`},
		{"err-struct-int", false, args{Empty{}, reflect.TypeOf(0)}, nil, `cannot convert from conv\.Empty to int`},
		{"err-struct-struct", false, args{Empty{}, reflect.TypeOf(Empty{})}, nil, `cannot convert from conv\.Empty to conv\.Empty`},
		{"err-cust", true, args{time.Unix(0, 0), reflect.TypeOf("")}, nil, "we make a custom error for zero time"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = customTimeConv.SimpleToSimple(tt.args.src, tt.args.dstType)
			} else {
				got, err = _defaultConv.SimpleToSimple(tt.args.src, tt.args.dstType)
			}

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("SimpleToSimple() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("SimpleToSimple() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimpleToSimple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_SliceToSlice(t *testing.T) {
	var nilI []int
	var nilStruct []struct{}

	type args struct {
		src         interface{}
		dstSliceTyp reflect.Type
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		errRegex string
	}{
		{"empty", args{[]int{}, reflect.TypeOf([]string{})}, []string{}, ""},
		{"int-int", args{[]int{1, 2, 3}, reflect.TypeOf([]int{})}, []int{1, 2, 3}, ""},
		{"int-float", args{[]int{1, 2, 3333333}, reflect.TypeOf([]float32{})}, []float32{1, 2, 3333333}, ""},
		{"string-int", args{[]string{"123", "321"}, reflect.TypeOf([]int{})}, []int{123, 321}, ""},
		{"string-bool", args{[]string{"true", "1", "0"}, reflect.TypeOf([]bool{})}, []bool{true, true, false}, ""},
		{"bool-string", args{[]bool{true, true, false}, reflect.TypeOf([]string{})}, []string{"1", "1", "0"}, ""},
		{"nil-nil", args{nilI, reflect.TypeOf([]struct{}{})}, nilStruct, ""},
		{"interface-interface", args{[]interface{}{1, "v"}, reflect.TypeOf([]interface{}{})}, []interface{}{1, "v"}, ""},

		{"err", args{[]struct{}{{}}, reflect.TypeOf([]string{})}, nil, "^conv.SliceToSlice: .+, at index 0.+"},
		{"err-nil", args{nil, reflect.TypeOf([]string{})}, nil, "should not be nil"},
		{"err-src", args{1, reflect.TypeOf([]string{})}, nil, "src must be a slice"},
		{"err-dst", args{[]int{1, 2, 3}, reflect.TypeOf(1)}, nil, "the destination type must be slice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _defaultConv.SliceToSlice(tt.args.src, tt.args.dstSliceTyp)

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("SliceToSlice() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("SliceToSlice() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_MapToStruct(t *testing.T) {
	type args struct {
		c        *Conv
		m        map[string]interface{}
		dstTyp   reflect.Type
		want     interface{}
		errRegex string
	}
	check := func(t *testing.T, args args) {
		got, err := args.c.MapToStruct(args.m, args.dstTyp)

		if err != nil {
			if args.errRegex == "" {
				t.Errorf("MapToStruct() unexpected error = %v", err)
			}

			if match, _ := regexp.MatchString(args.errRegex, err.Error()); !match {
				t.Errorf("MapToStruct() error = %v , must match %v",
					strconv.Quote(err.Error()), strconv.Quote(args.errRegex))
			}
		}

		if !reflect.DeepEqual(got, args.want) {
			t.Errorf("MapToStruct() = %v, want %v", got, args.want)
		}
	}

	t.Run("ok-match", func(t *testing.T) {
		type T struct {
			S     FromString
			I     int
			F     float64
			E     interface{}
			inner int
		}

		check(t, args{
			c:        _defaultConv,
			m:        map[string]interface{}{"I": 1, "F": 3.14, "S": "vv", "inner": 1, "E": nil},
			dstTyp:   reflect.TypeOf(T{}),
			want:     T{I: 1, F: 3.14, S: "vv", inner: 0},
			errRegex: "",
		})
	})

	t.Run("ok-mismatch", func(t *testing.T) {
		type T struct {
			S FromString
			I int
			F float64
			E interface{}
		}

		check(t, args{
			c:        _defaultConv,
			m:        map[string]interface{}{"I2": 1, "F2": 3.14, "S2": "vv", "E2": nil},
			dstTyp:   reflect.TypeOf(T{}),
			want:     T{},
			errRegex: "",
		})
	})

	t.Run("err-nil", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			m:        map[string]interface{}(nil),
			dstTyp:   reflect.TypeOf(struct{}{}),
			want:     nil,
			errRegex: "should not be nil",
		})
	})

	t.Run("err-type", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			m:        map[string]interface{}{},
			dstTyp:   reflect.TypeOf(1),
			want:     nil,
			errRegex: "type must be struct",
		})
	})

	t.Run("err-field", func(t *testing.T) {
		type T struct{ F float32 }

		check(t, args{
			c:        _defaultConv,
			m:        map[string]interface{}{"F": "x"},
			dstTyp:   reflect.TypeOf(T{}),
			want:     nil,
			errRegex: "error on converting field 'F': .+",
		})
	})

	t.Run("conv-basetype", func(t *testing.T) {
		type Base struct{ V string }
		type E Base
		type T struct {
			S FromString
			I FromInt
			E E
		}

		check(t, args{
			c: _defaultConv,
			m: map[string]interface{}{
				"S": "vv",
				"I": 12,
				"E": map[string]interface{}{
					"V": "x",
				},
			},
			dstTyp: reflect.TypeOf(T{}),
			want: T{
				I: 12,
				S: "vv",
				E: E{
					V: "x",
				},
			},
			errRegex: "",
		})
	})

	t.Run("conv-case-insensitive", func(t *testing.T) {
		type T struct {
			S string
			I int
			F float64
		}

		check(t, args{
			c: _caseInsensitiveConv,
			m: map[string]interface{}{
				"i": 1,
				"f": 3.14,
				"s": "vv",
			},
			dstTyp:   reflect.TypeOf(T{}),
			want:     T{I: 1, F: 3.14, S: "vv"},
			errRegex: "",
		})
	})

	t.Run("conv-tag", func(t *testing.T) {
		type T struct {
			S string  `conv:"str"`
			I int     `conv:"int"`
			F float64 `conv:"flt"`
		}

		check(t, args{
			c: _tagConv,
			m: map[string]interface{}{
				"int": 1,
				"flt": 3.14,
				"str": "vv",
			},
			dstTyp:   reflect.TypeOf(T{}),
			want:     T{I: 1, F: 3.14, S: "vv"},
			errRegex: "",
		})
	})

	t.Run("conv-embedded-struct", func(t *testing.T) {
		type E struct {
			V1, V2 string
		}
		type T struct {
			S string
			E
			I int
		}
		// Here T is equivalent to struct{S,V1,V2 string; I int}

		check(t, args{
			c: _defaultConv,
			m: map[string]interface{}{
				"I":  1,
				"S":  "vv",
				"V1": "v1",
				"V2": "v2",
			},
			dstTyp: reflect.TypeOf(T{}),
			want: T{
				I: 1,
				S: "vv",
				E: E{
					V1: "v1",
					V2: "v2",
				},
			},
			errRegex: "",
		})
	})

	t.Run("ok-embedded-struct-with-tag", func(t *testing.T) {
		type E struct {
			V1    string
			VE    string `conv:"e"`
			inner string `conv:"VE"` //lint:ignore U1000 Unexported field is ignored.
		}
		type T struct {
			S FromString   `conv:"str"`
			E `conv:"emb"` // E has tag, so it is treated as a non-embedded field.
		}

		check(t, args{
			c: _tagConv,
			m: map[string]interface{}{
				"emb": map[string]interface{}{
					"V1": "v1",
					"e":  "ve",
					"VE": "x", // Not processed.
				},
			},
			dstTyp: reflect.TypeOf(T{}),
			want: T{
				E: E{
					V1: "v1",
					VE: "ve",
				},
			},
			errRegex: "",
		})
	})
}

func TestConv_MapToMap(t *testing.T) {
	type args struct {
		m      interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		errRegex string
	}{
		{
			"nil",
			args{
				map[string]int(nil),
				reflect.TypeOf(map[float32]int(nil)),
			},
			map[float32]int(nil),
			"",
		},

		{
			"si-si",
			args{
				map[string]int{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				reflect.TypeOf(map[string]int(nil)),
			},
			map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			"",
		},

		{
			"si-is",
			args{
				map[string]int{
					"1":  11,
					"2":  22,
					"33": 3,
				},
				reflect.TypeOf(map[int]string(nil)),
			},
			map[int]string{
				1:  "11",
				2:  "22",
				33: "3",
			},
			"",
		},

		{
			"[]interface-interface",
			args{
				map[string]interface{}{
					"1": []interface{}{1, 2, 3},
				},
				reflect.TypeOf(map[interface{}]interface{}(nil)),
			},
			map[interface{}]interface{}{
				"1": []interface{}{1, 2, 3},
			},
			"",
		},

		{
			"err-src",
			args{
				123,
				reflect.TypeOf(map[int]string(nil)),
			},
			nil,
			"must be a map",
		},

		{
			"err-typ",
			args{
				map[string]int{},
				reflect.TypeOf(1),
			},
			nil,
			"destination type must be map",
		},

		{
			"err-key",
			args{
				map[string]int{"a": 1},
				reflect.TypeOf(map[int]int{}),
			},
			nil,
			"cannot covert key 'a' to int: .+",
		},

		{
			"err-elem",
			args{
				map[string]string{"aa": "x"},
				reflect.TypeOf(map[string]int{}),
			},
			nil,
			"cannot covert value of key 'aa' to int: .+",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _defaultConv.MapToMap(tt.args.m, tt.args.dstTyp)

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("MapToMap() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("MapToMap() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_StructToMap(t *testing.T) {
	type args struct {
		c        *Conv
		src      interface{}
		want     map[string]interface{}
		errRegex string
	}
	check := func(t *testing.T, args args) {
		got, err := args.c.StructToMap(args.src)

		if err != nil {
			if args.errRegex == "" {
				t.Errorf("StructToMap() unexpected error = %v", err)
			}

			if match, _ := regexp.MatchString(args.errRegex, err.Error()); !match {
				t.Errorf("StructToMap() error = %v , must match %v",
					strconv.Quote(err.Error()), strconv.Quote(args.errRegex))
			}
		} else if args.errRegex != "" {
			t.Errorf("StructToMap() want error, got nil, pattern = %v", args.errRegex)
		}

		if !reflect.DeepEqual(got, args.want) {
			t.Errorf("StructToMap() = %v, want %v", got, args.want)
		}
	}

	t.Run("nil", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      nil,
			want:     nil,
			errRegex: "^conv.StructToMap: .+should not be nil",
		})
	})

	t.Run("simple", func(t *testing.T) {
		now := time.Now()
		check(t, args{
			c: _defaultConv,
			src: struct {
				Str   string
				Flt   float64
				Time  time.Time
				inner int
			}{"aa", 0.5, now, 4},
			want: map[string]interface{}{
				"Str":  "aa",
				"Flt":  0.5,
				"Time": now,
			},
			errRegex: "",
		})
	})

	t.Run("field-map-slice-without-value", func(t *testing.T) {
		type T struct {
			MNil   map[string]int
			MEmpty map[string]int
			SNil   []struct{}
			SEmpty []struct{}
		}

		check(t, args{
			c: _defaultConv,
			src: T{
				MEmpty: map[string]int{},
				SEmpty: []struct{}{},
			},
			want: map[string]interface{}{
				"MNil":   map[string]interface{}(nil),
				"MEmpty": map[string]interface{}{},
				"SNil":   []map[string]interface{}(nil),
				"SEmpty": []map[string]interface{}{},
			},
			errRegex: "",
		})
	})

	t.Run("field-map-with-value", func(t *testing.T) {
		type T struct {
			M map[string]int
		}

		check(t, args{
			c: _defaultConv,
			src: T{
				M: map[string]int{"A": 1, "B": 2},
			},
			want: map[string]interface{}{
				"M": map[string]interface{}{
					"A": 1,
					"B": 2,
				},
			},
			errRegex: "",
		})
	})

	t.Run("field-slice-with-value", func(t *testing.T) {
		type Inner struct {
			A string
			B []byte
		}
		type T struct{ In []Inner }

		check(t, args{
			c: _defaultConv,
			src: T{
				In: []Inner{
					{"A1", []byte{1, 2}},
					{"A2", []byte{3, 4}},
				},
			},
			want: map[string]interface{}{
				"In": []map[string]interface{}{
					{"A": "A1", "B": []byte{1, 2}},
					{"A": "A2", "B": []byte{3, 4}},
				},
			},
			errRegex: ``,
		})
	})

	t.Run("err-src-kind", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      1,
			want:     nil,
			errRegex: "must be a struct",
		})
	})

	t.Run("err-field-not-simple", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      struct{ C chan int }{make(chan int)},
			want:     nil,
			errRegex: "^conv.StructToMap: error on converting field C: conversion not supported on type chan",
		})
	})

	t.Run("err-field-slice-elem-invalid", func(t *testing.T) {
		type T struct{ V []chan int }

		check(t, args{
			c:        _defaultConv,
			src:      T{[]chan int{}},
			want:     nil,
			errRegex: `cannot convert \[\]chan int`,
		})
	})

	t.Run("err-field-map-key-invalid", func(t *testing.T) {
		type T struct{ In map[chan int]int }

		check(t, args{
			c:        _defaultConv,
			src:      T{map[chan int]int{make(chan int): 1}},
			want:     nil,
			errRegex: `field In: key .+?: .+cannot convert chan int to string`,
		})
	})

	t.Run("err-field-map-value-invalid", func(t *testing.T) {
		type T struct{ In map[int]chan int }

		check(t, args{
			c:        _defaultConv,
			src:      T{map[int]chan int{13: make(chan int)}},
			want:     nil,
			errRegex: `field In: value of key 13: conversion not supported on type chan`,
		})
	})

	t.Run("pointer-nil", func(t *testing.T) {
		type T struct{ In *int }

		check(t, args{
			c:        _defaultConv,
			src:      T{},
			want:     map[string]interface{}{"In": nil},
			errRegex: ``,
		})
	})

	t.Run("pointer-value", func(t *testing.T) {
		type T struct{ In *struct{ A int } }

		check(t, args{
			c:   _defaultConv,
			src: T{&struct{ A int }{33}},
			want: map[string]interface{}{
				"In": map[string]interface{}{"A": 33},
			},
			errRegex: ``,
		})
	})

	t.Run("embedded-struct", func(t *testing.T) {
		type E2 struct {
			VV1, VV2 string
		}
		type E struct {
			E2
			V1 int
		}
		type T struct {
			S string
			E
		}

		check(t, args{
			c: _defaultConv,
			src: T{
				S: "ss",
				E: E{
					V1: 33,
					E2: E2{
						VV1: "vv1",
						VV2: "vv2",
					},
				},
			},
			want: map[string]interface{}{
				"S":   "ss",
				"V1":  33,
				"VV1": "vv1",
				"VV2": "vv2",
			},
			errRegex: ``,
		})
	})

	t.Run("embedded-struct-with-tag", func(t *testing.T) {
		type E2 struct {
			VV2 string `conv:"value2"`
		}
		type E struct {
			E2 `conv:"ee2"`
			V1 int `conv:"value1"`
		}
		type T struct {
			E `conv:"ee"`
		}

		// TODO Convert from struct fields with tags are not supported now, it should come in the future.
		check(t, args{
			c: _tagConv,
			src: T{
				E: E{
					V1: 12,
					E2: E2{
						VV2: "vv2",
					},
				},
			},
			want: map[string]interface{}{
				"V1":  12,
				"VV2": "vv2",
			},
			errRegex: ``,
		})
	})
}

func TestConv_StructToStruct(t *testing.T) {
	type args struct {
		c        *Conv
		src      interface{}
		dstTyp   reflect.Type
		want     interface{}
		errRegex string
	}
	check := func(t *testing.T, args args) {
		got, err := args.c.StructToStruct(args.src, args.dstTyp)

		if err != nil {
			if args.errRegex == "" {
				t.Errorf("unexpected error = %v", err)
			}

			if match, _ := regexp.MatchString(args.errRegex, err.Error()); !match {
				t.Errorf("error = %v , must match %v",
					strconv.Quote(err.Error()), strconv.Quote(args.errRegex))
			}
		} else if args.errRegex != "" {
			t.Errorf("want error, got nil, pattern = %v", args.errRegex)
		}

		if !reflect.DeepEqual(got, args.want) {
			t.Errorf("want %v, got %v", args.want, got)
		}
	}

	t.Run("err-nil", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      nil,
			dstTyp:   reflect.TypeOf(struct{}{}),
			want:     nil,
			errRegex: "^conv.StructToStruct: the source value should not be nil$",
		})
	})

	t.Run("err-src", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      1,
			dstTyp:   reflect.TypeOf(struct{}{}),
			want:     nil,
			errRegex: "^conv.StructToStruct: the given value must be a struct, got int$",
		})
	})

	t.Run("err-dst", func(t *testing.T) {
		check(t, args{
			c:        _defaultConv,
			src:      struct{}{},
			dstTyp:   reflect.TypeOf(1),
			want:     nil,
			errRegex: "^conv.StructToStruct: the destination type must be struct, got int$",
		})
	})

	t.Run("err-field-nil", func(t *testing.T) {
		type from struct {
			V interface{}
		}
		type to struct {
			V int
		}

		check(t, args{
			c:        _defaultConv,
			src:      from{},
			dstTyp:   reflect.TypeOf(to{}),
			want:     nil,
			errRegex: "^conv.StructToStruct: error on converting field V: conv.ConvertType: cannot convert nil to int$",
		})
	})

	t.Run("err-field-type", func(t *testing.T) {
		type from struct {
			V chan int
		}
		type to struct {
			V int
		}

		check(t, args{
			c:        _defaultConv,
			src:      from{V: make(chan int)},
			dstTyp:   reflect.TypeOf(to{}),
			want:     nil,
			errRegex: "^conv.StructToStruct: error on converting field V: conv.ConvertType: cannot convert chan int to int$",
		})
	})

	t.Run("field-mismatch", func(t *testing.T) {
		type from struct {
			V chan int // Mismatched field will not cause error.
		}
		type to struct {
			Other int
		}

		check(t, args{
			c:        _defaultConv,
			src:      from{V: make(chan int)},
			dstTyp:   reflect.TypeOf(to{}),
			want:     to{},
			errRegex: "",
		})
	})

	t.Run("clone", func(t *testing.T) {
		type T struct {
			Str   FromString
			Int   int
			Flt   float64
			Empty interface{}
			inner int
		}

		check(t, args{
			c:        _defaultConv,
			src:      T{Str: "gg", Int: 333, Flt: -1.23, inner: 44},
			dstTyp:   reflect.TypeOf(T{}),
			want:     T{Str: "gg", Int: 333, Flt: -1.23},
			errRegex: "",
		})
	})

	t.Run("clone-case-insensitive", func(t *testing.T) {
		type from struct {
			Out string
			out string //lint:ignore U1000 Ignored.
			Sl  []byte
		}
		type to struct {
			OUt float64
			SL  []int
		}

		check(t, args{
			c:        _caseInsensitiveConv,
			src:      from{Out: "-1999", Sl: []byte{3, 5, 77}},
			dstTyp:   reflect.TypeOf(to{}),
			want:     to{OUt: float64(-1999), SL: []int{3, 5, 77}},
			errRegex: "",
		})
	})

	t.Run("embedded-struct", func(t *testing.T) {
		type EFrom struct {
			V1 int
		}
		type from struct {
			EFrom
			V2 int
		}

		type ETo struct {
			V2 string
		}
		type to struct {
			V1 string
			ETo
		}

		check(t, args{
			c: _defaultConv,
			src: from{
				EFrom: EFrom{11},
				V2:    22,
			},
			dstTyp: reflect.TypeOf(to{}),
			want: to{
				V1:  "11",
				ETo: ETo{"22"},
			},
			errRegex: "",
		})
	})

	t.Run("embedded-struct-with-tag", func(t *testing.T) {
		type Ef struct {
			V1 int
		}
		type from struct {
			Ef `conv:"F"`
		}

		type Et struct {
			V int `conv:"V1"`
		}
		type to struct {
			F Et
		}

		// TODO Convert from struct fields with tags are not supported now, it should come in the future.
		check(t, args{
			c: _tagConv,
			src: from{
				Ef: Ef{V1: 33},
			},
			dstTyp:   reflect.TypeOf(to{}),
			want:     to{}, // Expect t{F: Et{V: 33}} when tags are supported.
			errRegex: "",
		})
	})
}

func TestConv_ConvertType_convertPointers(t *testing.T) {
	i := 1
	pi := &i
	ppi := &pi
	pppi := &ppi

	s := "3"
	ps := &s
	pps := &ps

	type args struct {
		src    interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name      string
		args      args
		wantVal   interface{}
		wantDepth int
		wantErr   bool
	}{
		{"i-i", args{i, reflect.TypeOf(i)}, i, 0, false},
		{"pi-i", args{pi, reflect.TypeOf(i)}, i, 0, false},
		{"ppi-i", args{ppi, reflect.TypeOf(i)}, i, 0, false},
		{"pppi-i", args{pppi, reflect.TypeOf(i)}, i, 0, false},

		{"i-pi", args{i, reflect.TypeOf(pi)}, i, 1, false},
		{"i-ppi", args{i, reflect.TypeOf(ppi)}, i, 2, false},
		{"i-pppi", args{i, reflect.TypeOf(pppi)}, i, 3, false},

		{"pi-pi", args{pi, reflect.TypeOf(pi)}, i, 1, false},
		{"ppi-ppi", args{ppi, reflect.TypeOf(ppi)}, i, 2, false},
		{"pppi-pppi", args{pppi, reflect.TypeOf(pppi)}, i, 3, false},

		{"ppi-pi", args{pi, reflect.TypeOf(pi)}, i, 1, false},
		{"pppi-pi", args{pppi, reflect.TypeOf(pi)}, i, 1, false},
		{"pi-ppi", args{ppi, reflect.TypeOf(ppi)}, i, 2, false},

		{"s-s", args{s, reflect.TypeOf(s)}, s, 0, false},
		{"ps-s", args{ps, reflect.TypeOf(s)}, s, 0, false},
		{"s-ps", args{s, reflect.TypeOf(ps)}, s, 1, false},

		{"pi-s", args{pi, reflect.TypeOf(s)}, "1", 0, false},
		{"ppi-ps", args{ppi, reflect.TypeOf(ps)}, "1", 1, false},
		{"pi-pps", args{pi, reflect.TypeOf(pps)}, "1", 2, false},

		{"ps-i", args{ps, reflect.TypeOf(i)}, 3, 0, false},
		{"pps-ppi", args{pps, reflect.TypeOf(ppi)}, 3, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _defaultConv.ConvertType(tt.args.src, tt.args.dstTyp)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var valGot interface{}
			var valDepth int
			switch v := got.(type) {
			case int:
				valGot, valDepth = v, 0
			case *int:
				valGot, valDepth = *v, 1
			case **int:
				valGot, valDepth = **v, 2
			case ***int:
				valGot, valDepth = ***v, 3
			case string:
				valGot, valDepth = v, 0
			case *string:
				valGot, valDepth = *v, 1
			case **string:
				valGot, valDepth = **v, 2
			case ***string:
				valGot, valDepth = ***v, 3
			}

			if valGot != tt.wantVal || valDepth != tt.wantDepth {
				t.Errorf("ConvertType() = val %v, depth %v, want val %v depth %v", valGot, valDepth, tt.wantVal, tt.wantDepth)
			}
		})
	}
}

func TestConv_ConvertType_mapToStructWithPointers(t *testing.T) {
	type T struct {
		S FromString
		I int
		F float64
	}
	type P struct {
		S1    *T
		inner *int
		Out   *FromString
		m     *map[string]int
		Sl    *[]byte
	}

	fieldS1 := T{S: "23", I: 33, F: 44}
	fieldOut := FromString("3.14")
	p2 := P{S1: &fieldS1, Out: &fieldOut}
	pp2 := &p2
	in := map[string]interface{}{
		"S1":    map[string]interface{}{"S": 23, "I": 33, "F": 44, "inner": 55},
		"Out":   3.14,
		"Sl":    nil,
		"inner": -1,
	}
	res, err := _defaultConv.ConvertType(in, reflect.TypeOf(pp2))
	if err != nil {
		t.Errorf("ConvertType: %s", err)
		return
	}

	// reflect.DeepEqual() doesn't compare the underlying values of pointers.
	// We compare the fields manually.
	out := *res.(*P)
	if out.inner != nil {
		t.Error("inner != nil")
		return
	}

	if out.m != nil {
		t.Error("m != nil")
		return
	}

	if out.Sl != nil {
		t.Error("Sl != nil")
		return
	}

	if out.Out == nil || *out.Out != "3.14" {
		t.Error("Out != 3.14")
		return
	}

	if out.S1 == nil {
		t.Error("S1 == nil")
		return
	}

	if !reflect.DeepEqual(*out.S1, fieldS1) {
		t.Errorf("S1: got %#v, want %#v", *out.S1, fieldS1)
		return
	}
}

func TestConv_ConvertType_sliceToSlice(t *testing.T) {
	type s struct {
		S     string
		I     int64
		F     float32
		inner int
	}
	type sPtr struct {
		S     *string
		I     **int64
		F     *float32
		inner *int
	}

	in := []*s{
		{S: "1", I: 3, inner: 5},
		{S: "2", F: 4},
	}

	dstTyp := reflect.TypeOf([]*sPtr{})
	out, err := _defaultConv.ConvertType(in, dstTyp)
	if err != nil {
		t.Errorf("err: %s", err.Error())
		return
	}

	ss, ok := out.([]*sPtr)
	if !ok {
		t.Errorf("wrong type: %v", reflect.TypeOf(out))
		return
	}

	if len(ss) != len(in) {
		t.Errorf("wrong length, got %v, want %v", len(ss), len(in))
		return
	}

	check := func(idx int, s string, i int64, f float32) {
		e := ss[idx]
		if e == nil {
			t.Errorf("nil at index %v", idx)
			return
		}

		if e.S == nil || *e.S != s {
			t.Errorf("unexpected value at index %v field S", idx)
			return
		}

		if e.I == nil || **e.I != i {
			t.Errorf("unexpected value at index %v field I", idx)
			return
		}

		if e.F == nil || *e.F != f {
			t.Errorf("unexpected value at index %v field F", idx)
			return
		}

		if e.inner != nil {
			t.Errorf("unexpected value at index %v field inner", idx)
			return
		}
	}

	check(0, "1", 3, 0)
	check(1, "2", 0, 4)
}

func TestConv_ConvertType_flatMap(t *testing.T) {

	t.Run("simple", func(t *testing.T) {
		src := map[string]interface{}{
			"": 87654321,
		}
		got, err := _defaultConv.ConvertType(src, reflect.TypeOf(0))

		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		if got != 87654321 {
			t.Errorf("want %v, got %v", 87654321, got)
		}
	})

	t.Run("map", func(t *testing.T) {
		type T struct {
			S FromString
			I int
			F float64
		}

		i := 1999
		var pf *float32

		src := map[string]interface{}{
			"": map[interface{}]interface{}{
				struct{ I *int }{&i}: nil,
				struct{}{}:           pf,
				struct{ S int }{123}: &pf,
			},
		}
		got, err := _defaultConv.ConvertType(src, reflect.TypeOf(map[T][]int{}))

		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		want := map[T][]int{
			{I: 1999}:  nil,
			{}:         nil,
			{S: "123"}: nil,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("want %v, got %v", 87654321, got)
		}
	})
}

func TestConv_ConvertType(t *testing.T) {
	now := time.Now()

	type args struct {
		src    interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		errRegex string
	}{
		// simple to simple
		{"string-int", args{"-387656", reflect.TypeOf(0)}, -387656, ""},
		{"time-uint", args{now, reflect.TypeOf(uint(0))}, uint(now.Unix()), ""},

		// string to slice
		{"string-[]byte", args{"233", reflect.TypeOf([]byte{})}, []byte{233}, ""},

		// struct to map
		{
			"struct-map",
			args{
				struct{ A, B interface{} }{},
				reflect.TypeOf(map[string]interface{}{}),
			},
			map[string]interface{}{
				"A": nil,
				"B": nil,
			},
			"",
		},
		{
			"err-struct-wrong-map",
			args{
				struct{}{},
				reflect.TypeOf(map[int]interface{}{}),
			},
			nil,
			`^conv.ConvertType: .+destination type must be map\[string\]interface\{\}, got map\[int\]interface.?\{\}$`,
		},

		// map to struct
		{
			"err-wrong-map-string",
			args{
				map[float32]interface{}{},
				reflect.TypeOf(struct{}{}),
			},
			nil,
			`^conv.ConvertType: .+the map must be map\[string\]interface\{\}, got map\[float32\]interface.?\{\}$`,
		},

		// to empty interface
		{
			"empty-interface",
			args{
				[]int{1, 2, 3},
				typEmptyInterface,
			},
			[]int{1, 2, 3},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _defaultConv.ConvertType(tt.args.src, tt.args.dstTyp)

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("ConvertType() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("ConvertType() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_Convert_panic(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			var err interface{}
			if err = recover(); err == nil {
				t.Fatalf("should panic an error")
			}

			const wantMsg = "conv.Convert: the destination value must be a pointer"
			if err.(error).Error() != wantMsg {
				t.Fatalf("should panic an error with message: '%v', got '%v'", wantMsg, err)
			}
		}()

		_defaultConv.Convert(nil, 0)
	})

	t.Run("uninitialized", func(t *testing.T) {
		defer func() {
			var err interface{}
			if err = recover(); err == nil {
				t.Fatalf("should panic an error")
			}

			const wantMsg = "conv.Convert: the pointer must be initialized"
			if err.(error).Error() != wantMsg {
				t.Fatalf("should panic an error with message: '%v', got '%v'", wantMsg, err)
			}
		}()

		var p *int
		_defaultConv.Convert("", p)
	})

	t.Run("ptr-to-nil-ptr", func(t *testing.T) {
		defer func() {
			var err interface{}
			if err = recover(); err == nil {
				t.Fatalf("should panic an error")
			}

			const wantMsg = "conv.Convert: the underlying pointer must be initialized"
			if err.(error).Error() != wantMsg {
				t.Fatalf("should panic an error with message: '%v', got '%v'", wantMsg, err)
			}
		}()

		var p *int
		_defaultConv.Convert("", &p)
	})
}

func TestConv_Convert_ptr(t *testing.T) {
	i := 1
	pi := &i
	ppi := &pi

	t.Run("nil", func(t *testing.T) {
		_defaultConv.Convert(nil, pi)
		if *pi != 1 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})

	t.Run("string-p-int", func(t *testing.T) {
		_defaultConv.Convert("-54321", pi)
		if *pi != -54321 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})

	t.Run("string-pp-int", func(t *testing.T) {
		_defaultConv.Convert("12345", ppi)
		if **ppi != 12345 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})
}

func TestConv_withCustomConverters(t *testing.T) {
	type Name struct{ FirstName, LastName string }
	namePtrTyp := reflect.TypeOf(&Name{})
	nameTyp := namePtrTyp.Elem()

	//  "A B" -> Name{A, B}.
	nameStructConverter := func(value interface{}, typ reflect.Type) (result interface{}, err error) {
		if typ != nameTyp {
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

	// string/Name/*Name -> *Name.
	namePointerConverter := func(value interface{}, typ reflect.Type) (result interface{}, err error) {
		if typ != namePtrTyp {
			return nil, nil
		}

		if _, ok := value.(string); ok {
			n, err := nameStructConverter(value, nameTyp)
			if err != nil {
				return nil, err
			}
			return &n, nil
		}

		if v, ok := value.(Name); ok {
			return &Name{v.FirstName, v.LastName}, nil
		}

		if v, ok := value.(*Name); ok {
			return &Name{v.FirstName, v.LastName}, nil
		}

		return nil, nil
	}

	c := &Conv{
		Conf: Config{
			CustomConverters: []ConvertFunc{nameStructConverter, namePointerConverter},
		},
	}

	t.Run("StringToName", func(t *testing.T) {
		var got Name
		err := c.Convert("John Doe", &got)
		if err != nil {
			t.Errorf("got error %s", err)
		}

		want := Name{"John", "Doe"}
		if got != want {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("NameToNamePtr", func(t *testing.T) {
		got, err := c.ConvertType(Name{"John", "Doe"}, reflect.TypeOf(&Name{}))
		if err != nil {
			t.Errorf("got error %s", err)
		}

		n := got.(*Name)
		want := &Name{"John", "Doe"}
		if n.FirstName != want.FirstName || n.LastName != want.LastName {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("error", func(t *testing.T) {
		var got Name
		err := c.Convert("John X Doe", &got)
		if err == nil {
			t.Errorf("want error")
		}

		want := "conv.Convert: converter[0]: bad name"
		if err.Error() != want {
			t.Errorf("want error %s, got %s", want, err)
		}
	})

	t.Run("default", func(t *testing.T) {
		var got string
		err := c.Convert("John Doe", &got)
		if err != nil {
			t.Errorf("got error %s", err)
		}

		want := "John Doe"
		if got != want {
			t.Errorf("want %v, got %v", want, got)
		}
	})
}

func TestConv_tryFlattenEmptyKeyMap(t *testing.T) {
	c := &Conv{}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"n0", args{nil}, nil},
		{"n1", args{1}, nil},
		{"n2", args{map[string]interface{}{}}, nil},
		{"n3", args{map[string]string{"": "123"}}, nil},
		{"n4", args{map[string]interface{}{"": "123", "other": "a"}}, nil},
		{"y1", args{map[string]interface{}{"": "123"}}, "123"},
		{"y1", args{map[string]interface{}{"": []int{1, 2, 3}}}, []int{1, 2, 3}},
		{"v", args{map[string]interface{}{"other": "a"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.tryFlattenEmptyKeyMap(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tryFlattenEmptyKeyMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
