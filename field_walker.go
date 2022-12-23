package conv

import (
	"reflect"
	"sync"
)

var fieldWalkerCache syncMap

// FieldWalker is used to traverse all field of a struct.
//
// The traverse will go into each level of embedded and untagged structs. Unexported fields are ignored.
// It reads fields in this order:
//   - Tagged fields.
//   - Non-embedded struct or non-struct fields.
//   - Fields of embedded struct, recursively.
//
// e.g. without tagName
//
//	type Ec struct {
//	    D int
//	}
//	type Eb struct {
//	    B int // hided by T.B
//	    Ec    // hided by T.Ec.D
//	    C int
//	}
//	type T struct {
//	    A  int
//	    Eb
//	    B  string // hides Eb.B
//	    Ec
//	}
//
// The traverse order is:
//
//	PATH     INDEX
//	A        {0}
//	B        {2}
//	Eb.C     {1, 2}
//	Ec.D     {3, 0}
//
// e.g. with tagName="json"
//
//	type A struct {
//	  A int
//	  X int  // hided by T.B
//	}
//	type B struct {
//	  B1 int // absent
//	  B2 int // absent
//	}
//	type T struct {
//	  A
//	  B `json:"X"` // hides B.X, the traverse will not go into the field
//	}
//
// The traverse order is:
//
//	PATH     INDEX    TAG
//	B        {1}      X
//	A.A      {0, 0}
type FieldWalker struct {
	typ     reflect.Type
	tagName string
	mu      sync.Mutex
	fields  []FieldInfo
}

// FieldInfo describes a field in a struct.
type FieldInfo struct {
	reflect.StructField

	// If the field is a field of am embedded and untagged field which type is struct,
	// the path is a dot-split string like A.B.C; otherwise it's equal to F.Name.
	Path string

	// The tag value of the field.
	TagValue string
}

// NewFieldWalker creates a new instance of FieldWalker.
// When tagName is specified, the values of the tag will be filled into FieldInfo.TagValue during the traversal.
func NewFieldWalker(typ reflect.Type, tagName string) *FieldWalker {
	type key struct {
		reflect.Type
		string
	}
	v, _ := fieldWalkerCache.LoadOrStore(key{typ, tagName}, &FieldWalker{
		typ:     typ,
		tagName: tagName,
	})
	return v.(*FieldWalker)
}

// WalkFields walks through fields of the given type of struct (or pointer) with a breadth-first traverse.
// Each field will be send to the callback function. If the function returns false, the traverse stops.
func (walker *FieldWalker) WalkFields(callback func(FieldInfo) bool) {
	if walker.fields == nil {
		walker.initFields()
	}

	for _, fieldInfo := range walker.fields {
		if !callback(fieldInfo) {
			break
		}
	}
}

// WalkValues is like WalkFields(), but walks through all field values.
//
// If a struct is embedded as a pointer, and the value is nil, the field is ignored.
// If the given value is nil, the traverse stops with no callback.
func (walker *FieldWalker) WalkValues(value reflect.Value, callback func(FieldInfo, reflect.Value) bool) {
	if walker.fields == nil {
		walker.initFields()
	}

	// Try extract the underlying type of a pointer, stop on nil.
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}

		for {
			value = value.Elem()
			if value.Kind() != reflect.Ptr {
				break
			}
		}
	}

	for _, fieldInfo := range walker.fields {
		index := fieldInfo.Index
		embedded := fieldInfo.TagValue == "" && len(index) > 1

		v := value
		for i := 0; i < len(index); i++ {
			v = v.Field(index[i])

			if embedded {
				for v.Kind() == reflect.Ptr {
					if v.IsNil() {
						goto next
					}

					v = v.Elem()
				}
			}
		}

		if !callback(fieldInfo, v) {
			break
		}

	next:
	}
}

func (walker *FieldWalker) initFields() {
	walker.mu.Lock()
	defer walker.mu.Unlock()

	// Double-lock checking.
	if walker.fields != nil {
		return
	}

	fields := make([]FieldInfo, 0)
	visited := make(map[string]struct{})

	type fieldBuf struct {
		Index []int        // If the current field is an embedded field, stores the field index sequence.
		Path  string       // The field path, split by dots.
		Type  reflect.Type // The type of the current field.
	}

	// Dequeue and traverse the first element, enqueue the types of embedded structs, then return then new q.
	traverseOne := func(q []fieldBuf) []fieldBuf {
		buf, q := q[0], q[1:] // Dequeue.
		num := buf.Type.NumField()
		tagged := make([]bool, num)

		// Firstly read all tagged fields.
		if walker.tagName != "" {
			for i := 0; i < num; i++ {
				f := buf.Type.Field(i)

				// Ignore unexported fields. The document of PkgPath field says:
				// PkgPath is the package path that qualifies a lower case (unexported)
				// field name. It is empty for upper case (exported) field names.
				if len(f.PkgPath) > 0 {
					continue
				}

				tag := f.Tag.Get(walker.tagName)
				if tag == "" {
					continue
				}

				tagged[i] = true
				visited[tag] = struct{}{}

				fields = append(fields, FieldInfo{
					StructField: f,
					Path:        f.Name,
					TagValue:    tag,
				})
			}
		}

		// Read untagged fields.
		for i := 0; i < num; i++ {
			if tagged[i] {
				continue
			}

			f := buf.Type.Field(i)
			if len(f.PkgPath) > 0 {
				continue
			}

			if _, ok := visited[f.Name]; ok {
				continue
			}

			// Build the index sequence and path.
			f.Index = append(buf.Index, f.Index...)
			path := buf.Path
			if path != "" {
				path += "."
			}
			path += f.Name

			if f.Anonymous {
				// Try to extract the underlying type of a pointer.
				ft := f.Type
				for ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}

				// In a breadth-first traversal, the traverse of the embedded struct should be delayed.
				if ft.Kind() == reflect.Struct {
					q = append(q, fieldBuf{f.Index, path, ft})
					continue
				}
			}

			visited[f.Name] = struct{}{}
			fields = append(fields, FieldInfo{
				StructField: f,
				Path:        path,
			})
		}
		return q
	}

	typ := walker.typ
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	q := []fieldBuf{{Type: typ}}
	for {
		q = traverseOne(q)
		if len(q) == 0 {
			break
		}
	}

	walker.fields = fields
}
