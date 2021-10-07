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

type S1 struct {
	S     string
	I     int
	F     float64
	inner int
}

type S2 struct {
	S1
	inner int
	Out   string
	m     map[string]int
	Sl    []byte
}

type S3 struct {
	In    interface{}
	S1    interface{}
	S2    interface{}
	inner interface{}
}

type P2 struct {
	*S1
	inner *int
	Out   *string
	m     *map[string]int
	Sl    *[]byte
}

func TestConv_StringToSlice(t *testing.T) {
	customConv := Conv{
		SplitStringHelper: func(v string) []string { return strings.Split(v, "~") },
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
				got, err = Conv{}.StringToSlice(tt.args.v, tt.args.simpleSliceType)
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
			got, err := Conv{}.SimpleToBool(tt.args.v)
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
	customTimeConv := Conv{
		TimeToStringHelper: func(t time.Time) (string, error) {
			if t == time.Unix(0, 0) {
				return "", errors.New("we make a custom error for zero time")
			}
			return t.Format("1-02,2006 15:04:05.000"), nil
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
		{"float", false, args{3.14}, "3.14", false},
		{"string", false, args{"中"}, "中", false},
		{"time0", false, args{time.Date(2020, 1, 20, 13, 6, 22, int(321*time.Millisecond), time.UTC)}, "2020-01-20T13:06:22Z", false},
		{"time+5", false, args{time.Date(2020, 1, 20, 13, 6, 22, int(321*time.Millisecond), time.FixedZone("", 5*3600))}, "2020-01-20T13:06:22+05:00", false},

		{"nil", false, args{nil}, "", true},
		{"err1", false, args{struct{}{}}, "", true},
		{"err2", false, args{map[string]interface{}{}}, "", true},
		{"err-cust", true, args{time.Unix(0, 0)}, "", true},

		// Customized Conv.TimeToStringHelper() .
		{"cus-time", true, args{time.Date(2020, 1, 6, 13, 6, 22, int(321*time.Millisecond), time.UTC)}, "1-06,2020 13:06:22.321", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = customTimeConv.SimpleToString(tt.args.v)
			} else {
				got, err = Conv{}.SimpleToString(tt.args.v)
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
	spDate := time.Date(2021, 6, 3, 13, 21, 22, 54321, time.UTC).Local()
	spDateWithoutNano := time.Unix(spDate.Unix(), 0).Local()
	customTimeConv := Conv{
		StringToTimeHelper: func(v string) (time.Time, error) { return spDate, nil },
		TimeToStringHelper: func(t time.Time) (string, error) {
			if t == time.Unix(0, 0) {
				return "", errors.New("we make a custom error for zero time")
			}
			return t.Format("20060102"), nil
		},
	}

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
		{"utc-local", false, args{spDate.UTC(), reflect.TypeOf(spDate)}, spDate, ""},
		{"time-string", false, args{spDate, reflect.TypeOf("")}, spDate.Format(time.RFC3339), ""},
		{"time-string-custom", true, args{spDate, reflect.TypeOf("")}, "20210603", ""},
		{"string-time", false, args{"2021-06-03T13:21:22Z", reflect.TypeOf(spDate)}, spDateWithoutNano, ""},
		{"string-time-custom", true, args{"any", reflect.TypeOf(spDate)}, spDate, ""}, // always returns spDate
		{"time-int", false, args{spDate, reflect.TypeOf(0)}, int(spDate.Unix()), ""},
		{"time-float", false, args{spDate, reflect.TypeOf(0.0)}, float64(spDate.Unix()), ""},
		{"int-time", false, args{1622726482, reflect.TypeOf(spDate)}, spDateWithoutNano, ""},

		// err
		{"err-nil", false, args{nil, reflect.TypeOf(1)}, nil, "^conv.SimpleToSimple: the source value should not be nil$"},
		{"err-time-from-string", false, args{"date", reflect.TypeOf(spDate)}, nil, "^conv.SimpleToSimple: .+"},
		{"err-time-from-complex", false, args{1 + 3i, reflect.TypeOf(spDate)}, nil, "lost imaginary part"},
		{"err-time-to-int8", false, args{spDate, reflect.TypeOf(int8(0))}, nil, `value overflow`},
		{"err-struct-int", false, args{S1{}, reflect.TypeOf(0)}, nil, `cannot convert from conv\.S1 to int`},
		{"err-struct-struct", false, args{S1{}, reflect.TypeOf(S1{})}, nil, `cannot convert from conv\.S1 to conv\.S1`},
		{"err-cust", true, args{time.Unix(0, 0), reflect.TypeOf("")}, nil, "we make a custom error for zero time"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = customTimeConv.SimpleToSimple(tt.args.src, tt.args.dstType)
			} else {
				got, err = Conv{}.SimpleToSimple(tt.args.src, tt.args.dstType)
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
	var nilS1 []S1

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
		{"nil-nil", args{nilI, reflect.TypeOf([]S1{})}, nilS1, ""},

		{"err", args{[]struct{}{{}}, reflect.TypeOf([]string{})}, nil, "^conv.SliceToSlice: .+, at index 0.+"},
		{"err-nil", args{nil, reflect.TypeOf([]string{})}, nil, "should not be nil"},
		{"err-src", args{1, reflect.TypeOf([]string{})}, nil, "src must be a slice"},
		{"err-dst", args{[]int{1, 2, 3}, reflect.TypeOf(1)}, nil, "the destination type must be slice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.SliceToSlice(tt.args.src, tt.args.dstSliceTyp)

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
	caseInsensitiveConv := Conv{IndexNameHelper: CaseInsensitiveIndexName}

	type args struct {
		m      map[string]interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name        string
		useCustConv bool
		args        args
		want        interface{}
		errRegex    string
	}{
		{
			"ok1",
			false,
			args{
				map[string]interface{}{"I": 1, "F": 3.14, "S": "vv", "inner": 1},
				reflect.TypeOf(S1{}),
			},
			S1{I: 1, F: 3.14, S: "vv", inner: 0},
			"",
		},

		{
			"ok2",
			false,
			args{
				map[string]interface{}{"I2": 1, "F": 3.14, "S2": "vv"},
				reflect.TypeOf(S1{}),
			},
			S1{F: 3.14},
			"",
		},

		{
			"nil",
			false,
			args{
				map[string]interface{}(nil),
				reflect.TypeOf(S1{}),
			},
			nil,
			"should not be nil",
		},

		{
			"err-type",
			false,
			args{
				map[string]interface{}{"i": 1, "f": 3.14, "s": "vv"},
				reflect.TypeOf(1),
			},
			nil,
			"type must be struct",
		},

		{
			"err-field",
			false,
			args{
				map[string]interface{}{"i": 1, "F": "x", "s": "vv"},
				reflect.TypeOf(S1{}),
			},
			nil,
			"error on converting field 'F': .+",
		},

		// Test custom IndexNameHelper.
		{
			"ok1-ci",
			true,
			args{
				map[string]interface{}{"i": 1, "f": 3.14, "s": "vv"},
				reflect.TypeOf(S1{}),
			},
			S1{I: 1, F: 3.14, S: "vv"},
			"",
		},

		{
			"ok2-ci",
			true,
			args{
				map[string]interface{}{},
				reflect.TypeOf(S1{}),
			},
			S1{},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = caseInsensitiveConv.MapToStruct(tt.args.m, tt.args.dstTyp)
			} else {
				got, err = Conv{}.MapToStruct(tt.args.m, tt.args.dstTyp)
			}

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("MapToStruct() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("MapToStruct() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
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
			got, err := Conv{}.MapToMap(tt.args.m, tt.args.dstTyp)

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
		v interface{}
	}
	tests := []struct {
		name     string
		args     args
		want     map[string]interface{}
		errRegex string
	}{
		{
			"nil",
			args{nil},
			nil,
			"^conv.StructToMap: .+should not be nil",
		},

		{
			"simple1",
			args{
				S1{S: "1", F: 3, inner: 4},
			},
			map[string]interface{}{
				"S": "1",
				"I": 0,
				"F": float64(3),
			},
			"",
		},

		{
			"s2-empty",
			args{
				S2{},
			},
			map[string]interface{}{
				"S1":  map[string]interface{}{"S": "", "I": 0, "F": float64(0)},
				"Out": "",
				"Sl":  []byte(nil),
			},
			"",
		},

		{
			"s2",
			args{
				S2{
					S1:    S1{S: "1", F: 3.3, inner: 4},
					Out:   "o",
					Sl:    []byte{100, 255},
					m:     map[string]int{"a": 1},
					inner: 33,
				},
			},
			map[string]interface{}{
				"S1": map[string]interface{}{
					"S": "1",
					"I": 0,
					"F": float64(3.3),
				},
				"Out": "o",
				"Sl":  []byte{100, 255},
			},
			"",
		},

		{
			"err-src-kind",
			args{
				1,
			},
			nil,
			"must be a struct",
		},

		{
			"err-field-not-simple",
			args{S3{In: make(chan int)}},
			nil,
			"^conv.StructToMap: error on converting field In: must be a simple type, got chan$",
		},

		{
			"err-slice-elem-not-supported",
			args{S3{In: []chan int{}}},
			nil,
			`cannot convert \[\]chan int`,
		},

		{
			"err-map-key",
			args{S3{In: map[chan int]int{make(chan int): 1}}},
			nil,
			`field In: key .+?: .+cannot convert chan int to string`,
		},

		{
			"err-map-value",
			args{S3{In: map[int]chan int{13: make(chan int)}}},
			nil,
			`field In: value of key 13: must be a simple type, got chan`,
		},

		{
			"s3-map",
			args{S3{In: map[int]string{1: "a", 2: "b"}}},
			map[string]interface{}{"In": map[string]interface{}{"1": "a", "2": "b"}},
			"",
		},

		{
			"s3-map-nil",
			args{S3{In: map[int]string(nil)}},
			map[string]interface{}{"In": map[string]interface{}(nil)},
			"",
		},

		{
			"s3-slice-empty",
			args{S3{In: []S1{}, inner: 3}},
			map[string]interface{}{"In": []map[string]interface{}{}},
			"",
		},

		{
			"s3-slice-nil",
			args{S3{In: []S1(nil), inner: 3}},
			map[string]interface{}{"In": []map[string]interface{}(nil)},
			"",
		},

		{
			"s3-slice-struct",
			args{
				S3{
					In: []S2{
						{},
						{Out: "o", Sl: []byte{1, 2, 3}, S1: S1{S: "酷", F: 2}},
					},
				},
			},
			map[string]interface{}{
				"In": []map[string]interface{}{
					{"Out": "", "Sl": []byte(nil), "S1": map[string]interface{}{"S": "", "I": 0, "F": float64(0)}},
					{"Out": "o", "Sl": []byte{1, 2, 3}, "S1": map[string]interface{}{"S": "酷", "I": 0, "F": float64(2)}},
				},
			},
			"",
		},

		{
			"s3-slice-slice-empty",
			args{S3{In: [][]S1{}}},
			map[string]interface{}{"In": [][]map[string]interface{}{}},
			"",
		},

		{
			"s3-slice-slice",
			args{
				S3{
					In: [][]S1{
						{{S: "aa"}},
						{{I: 33, F: 1.1}},
					},
				},
			},
			map[string]interface{}{
				"In": [][]map[string]interface{}{
					{{"S": "aa", "I": 0, "F": float64(0)}},
					{{"S": "", "I": 33, "F": float64(1.1)}},
				},
			},
			"",
		},

		{
			"p2-nil-pointer",
			args{
				P2{},
			},
			map[string]interface{}{},
			"",
		},

		{
			"p2-non-nil-pointer",
			args{
				P2{
					S1:  &S1{},
					Out: new(string),
					Sl:  new([]byte),
				},
			},
			map[string]interface{}{
				"S1": map[string]interface{}{
					"S": "",
					"I": 0,
					"F": float64(0),
				},
				"Out": "",
				"Sl":  []byte(nil),
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.StructToMap(tt.args.v)

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("StructToMap() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("StructToMap() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_StructToStruct(t *testing.T) {
	caseInsensitiveConv := Conv{IndexNameHelper: CaseInsensitiveIndexName}

	type ss2 struct {
		OUt float64
		SL  []int
	}

	type args struct {
		src    interface{}
		dstTyp reflect.Type
	}
	tests := []struct {
		name        string
		useCustConv bool
		args        args
		want        interface{}
		errRegex    string
	}{
		{
			"err-nil",
			false,
			args{nil, reflect.TypeOf(S1{})},
			nil,
			"^conv.StructToStruct: the source value should not be nil$",
		},

		{
			"err-src",
			false,
			args{1, reflect.TypeOf(S1{})},
			nil,
			"must be a struct, got int",
		},

		{
			"err-dst",
			false,
			args{S1{}, reflect.TypeOf(1)},
			nil,
			"must be struct, got int",
		},

		{
			"err-field-nil",
			false,
			args{S3{}, reflect.TypeOf(S2{})},
			nil,
			"^conv.StructToStruct: error on converting field S1: conv.ConvertType: cannot convert nil to conv.S1$",
		},

		{
			"err-field-type",
			false,
			args{S3{S1: make(chan int)}, reflect.TypeOf(S2{})},
			nil,
			"^conv.StructToStruct: error on converting field S1: conv.ConvertType: cannot convert chan int to conv.S1$",
		},

		{
			"field-mismatch",
			false,
			args{S3{In: make(chan int), S1: S1{}}, reflect.TypeOf(S2{})},
			S2{},
			"",
		},

		{
			"clone",
			false,
			args{S1{S: "gg", I: 333, F: -1.23, inner: 44}, reflect.TypeOf(S1{})},
			S1{S: "gg", I: 333, F: -1.23},
			"",
		},

		// Test custom IndexNameHelper .
		{
			"custom-name-indexer",
			true,
			args{S2{Out: "-1999", Sl: []byte{3, 5, 77}, inner: 44}, reflect.TypeOf(ss2{})},
			ss2{OUt: float64(-1999), SL: []int{3, 5, 77}},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.useCustConv {
				got, err = caseInsensitiveConv.StructToStruct(tt.args.src, tt.args.dstTyp)
			} else {
				got, err = Conv{}.StructToStruct(tt.args.src, tt.args.dstTyp)
			}

			if err != nil {
				if tt.errRegex == "" {
					t.Errorf("StructToMap() unexpected error = %v", err)
				}

				if match, _ := regexp.MatchString(tt.errRegex, err.Error()); !match {
					t.Errorf("StructToMap() error = %v , must match %v",
						strconv.Quote(err.Error()), strconv.Quote(tt.errRegex))
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConv_ConvertType_ConvertPointers(t *testing.T) {
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
			got, err := Conv{}.ConvertType(tt.args.src, tt.args.dstTyp)

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

func TestConv_ConvertType_MapToStructWithPointers(t *testing.T) {
	fieldS1 := S1{S: "23", I: 33, F: 44}
	fieldOut := "3.14"
	p2 := P2{S1: &fieldS1, Out: &fieldOut}
	pp2 := &p2
	in := map[string]interface{}{
		"S1":    map[string]interface{}{"S": 23, "I": 33, "F": 44, "inner": 55},
		"Out":   3.14,
		"Sl":    nil,
		"inner": -1,
	}
	res, err := Conv{}.ConvertType(in, reflect.TypeOf(pp2))
	if err != nil {
		t.Errorf("ConvertType: %s", err)
		return
	}

	// reflect.DeepEqual() doesn't compare the underlying values of pointers.
	// We compare the fields manually.
	out := *res.(*P2)
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

func TestConv_ConvertType_SliceToSlice(t *testing.T) {
	type s struct {
		S     *string
		I     **int64
		F     *float32
		inner *int
	}

	in := []*S1{
		{S: "1", I: 3, inner: 5},
		{S: "2", F: 4},
	}

	dstTyp := reflect.TypeOf([]*s{})
	out, err := Conv{}.ConvertType(in, dstTyp)
	if err != nil {
		t.Errorf("err: %s", err.Error())
		return
	}

	ss, ok := out.([]*s)
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

func TestConv_ConvertType_FlatternMap(t *testing.T) {
	i := 1999
	var pf *float32

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
		{
			"simple",
			args{
				map[string]interface{}{
					"": 87654321,
				},
				reflect.TypeOf(0),
			},
			87654321,
			"",
		},

		{
			"map[S1][]int",
			args{
				map[string]interface{}{
					"": map[interface{}]interface{}{
						struct{ I *int }{&i}: nil,
						struct{}{}:           pf,
						struct{ S int }{123}: &pf,
					},
				},
				reflect.TypeOf(map[S1][]int{}),
			},
			map[S1][]int{
				{I: 1999}:  nil,
				{}:         nil,
				{S: "123"}: nil,
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.ConvertType(tt.args.src, tt.args.dstTyp)

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
				S3{},
				reflect.TypeOf(map[string]interface{}{}),
			},
			map[string]interface{}{},
			"",
		},
		{
			"err-struct-wrong-map",
			args{
				S1{},
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
				reflect.TypeOf(S1{}),
			},
			nil,
			`^conv.ConvertType: .+the map must be map\[string\]interface\{\}, got map\[float32\]interface.?\{\}$`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Conv{}.ConvertType(tt.args.src, tt.args.dstTyp)

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

func TestConv_Convert_PanicOnNil(t *testing.T) {
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

	Conv{}.Convert(nil, 0)
}

func TestConv_Convert_PanicOnUninitialized(t *testing.T) {
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
	Conv{}.Convert("", p)
}

func TestConv_Convert_Ptr(t *testing.T) {
	i := 1
	pi := &i
	ppi := &pi

	t.Run("nil", func(t *testing.T) {
		Conv{}.Convert(nil, pi)
		if *pi != 1 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})

	t.Run("string-p-int", func(t *testing.T) {
		Conv{}.Convert("-54321", pi)
		if *pi != -54321 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})

	t.Run("string-pp-int", func(t *testing.T) {
		Conv{}.Convert("12345", ppi)
		if **ppi != 12345 {
			t.Errorf("want %v, got %v", i, *pi)
		}
	})
}

func TestConv_tryFlattenEmptyKeyMap(t *testing.T) {
	c := Conv{}

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
