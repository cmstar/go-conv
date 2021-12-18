package conv

import (
	"reflect"
	"strings"
	"sync"
	"unicode"
)

/*
Define FieldMatcher and provide some built-in implementation.
*/

// FieldMatcherCreator is used to create FieldMatcher instances when converting from map to struct or
// from struct to struct.
type FieldMatcherCreator interface {
	// GetMatcher returns a FieldMatcher for the given type of struct.
	// If the type is not Struct, it panics.
	GetMatcher(typ reflect.Type) FieldMatcher
}

// FieldMatcher is used to match names when converting from map to struct or from struct to struct.
type FieldMatcher interface {
	// MatchField returns the first matched field for the given name;
	// if no name can match, returns a zero value and false.
	// The field returned must be an exported field.
	MatchField(name string) (field reflect.StructField, ok bool)
}

// SimpleMatcherConfig configures SimpleMatcherCreator.
type SimpleMatcherConfig struct {
	// Tag specifies the tag name for the fields. When a name is given by the tag, the matcher
	// matches the field using the given name; otherwise the raw field name is used.
	// The tag works like some commonly used tags such as 'json', 'xml'.
	// Tag can be empty.
	//
	// e.g. When 'conv' is given as the tag name, the matcher returns the OldName field when
	// indexing 'NewName'.
	//   type Target struct {
	//       OldName int `conv:"NewName"` // 'NewName' can match this field.
	//       RawName                      // No tag specified, use 'RawName' for field matching.
	//   }
	//
	Tag string

	// CaseInsensitive specifies whether the matcher matches field names in a case-insensitive manner.
	// If this field is true, CamelSnakeCase is ignored.
	//
	// If the field is true, 'ab', 'Ab', 'aB', 'AB' are equal.
	//
	CaseInsensitive bool

	// OmitUnderscore specifies whether to omit underscores in field names.
	// If this field is true, CamelSnakeCase is ignored.
	//
	// If the field is true, 'ab', 'a_b', '_ab', 'a__b_' are equal.
	//
	OmitUnderscore bool

	// CamelSnakeCase whether to support camel-case and snake-case name comparing.
	// If CaseInsensitive or OmitUnderscore is true, this field is ignored.
	//
	// When it is set to true, the matcher can match names in camel-case or snake-case form, such as
	// 'lowerCaseCamel' or 'UpperCaseCamel' or 'snake_case' or 'Snake_Case' (sometimes called train-case).
	//
	// The first rune of each word is compared case-insensitively; others are compared in the case-sensitive
	// manner. For example:
	//   - These names are equal: aaBB, AaBb, aa_bb, Aa_Bb, aa_BB
	//   - These names are not equal: aa_bb, Aabb, aabB, _aaBb, AaBb_, Aa__Bb
	//
	// Mostly this option can be used to match field names that come from different platform, e.g.,
	// 'lowerCaseCamel' from Javascript, 'UpperCaseCamel' from Go, 'snake_case' from Mysql database.
	//
	CamelSnakeCase bool
}

// SimpleMatcherCreator returns an instance of FieldMatcherCreator.
// It is used as the default value when Conv.Config.FieldMatcher is nil.
type SimpleMatcherCreator struct {
	Conf SimpleMatcherConfig
	m    sync.Map
}

// GetMatcher implements FieldMatcherCreator.GetMatcher().
func (c *SimpleMatcherCreator) GetMatcher(typ reflect.Type) FieldMatcher {
	v, _ := c.m.LoadOrStore(typ, &simpleMatcher{
		conf: c.Conf,
		typ:  typ,
	})
	return v.(*simpleMatcher)
}

// simpleMatcher is the FieldMatcher returned by SimpleMatcherCreator.
type simpleMatcher struct {
	conf SimpleMatcherConfig // Conf configures the matcher.
	typ  reflect.Type        // The type of the struct.
	fs   *sync.Map           // The fields. A thread-safe map[string]reflect.StructField.
	mu   sync.Mutex          // Used to initialize fs.
}

func (ix *simpleMatcher) MatchField(name string) (reflect.StructField, bool) {
	// Init field mapping with double-lock check.
	// mu is used only to initialize fs, fs is sync.Map so it doesn't need another lock.
	if ix.fs == nil {
		ix.mu.Lock()
		if ix.fs == nil {
			ix.initFieldMap()
		}
		ix.mu.Unlock()
	}

	name = ix.fixName(name)
	if f, ok := ix.fs.Load(name); ok {
		return f.(reflect.StructField), ok
	}
	return reflect.StructField{}, false
}

func (ix *simpleMatcher) initFieldMap() {
	m := new(sync.Map)
	num := ix.typ.NumField()
	for i := 0; i < num; i++ {
		f := ix.typ.Field(i)

		// Ignore unexported fields. The document of PkgPath field says:
		// PkgPath is the package path that qualifies a lower case (unexported)
		// field name. It is empty for upper case (exported) field names.
		if len(f.PkgPath) > 0 {
			continue
		}

		// If a tag name is specified, use it; otherwise, use the raw field name.
		// TODO Consider process fields of embedded structs.
		var name string
		if ix.conf.Tag != "" {
			name = f.Tag.Get(ix.conf.Tag)
		}

		if name == "" {
			name = f.Name
		}
		name = ix.fixName(name)

		// As FieldMatcher.IndexName() says, it returns the first matched name,
		// When two field named may be transformed to the same name, we keep the first one.
		if _, ok := m.Load(name); ok {
			continue
		}

		m.Store(name, f)
	}
	ix.fs = m
}

func (ix *simpleMatcher) fixName(name string) string {
	supportCamel := true

	if ix.conf.CaseInsensitive {
		name = strings.ToLower(name)
		supportCamel = false
	}

	if ix.conf.OmitUnderscore {
		name = strings.Replace(name, "_", "", -1)
		supportCamel = false
	}

	if supportCamel && ix.conf.CamelSnakeCase {
		name = ix.fixCamelSnakeCaseName([]rune(name))
	}

	return name
}

// fixCamelSnakeCaseName transforms first runes of each word to '_c' format, 'c' is the rune in lower-case. e.g.:
//   aaBB   -> _aa_b_b
//   AaBb   -> _aa_bb
//   _a_b_  -> __a_b_
//
// c is the first rune of a word if any of:
// Case 1: The first rune of the name.
// Case 2: An uppercase rune.
// Case 3: A rune after a *single* underscore, and the underscore is not the first rune of a word.
//
func (ix *simpleMatcher) fixCamelSnakeCaseName(name []rune) string {
	var b strings.Builder
	b.Grow(len(name))

	const (
		sWordStart    byte = 's' // The first rune of a word.
		sDelimiter    byte = 'd' // A _ as a delimiter.
		sNonDelimiter byte = 'n' // A non-delimiter rune.
	)
	state := sWordStart

	for i := 0; i < len(name); i++ {
		c := name[i]

		// Case 1 & 2.
		if i == 0 || unicode.IsUpper(c) {
			state = sWordStart
			goto ensured
		}

		// Case 3.
		if state == sDelimiter {
			state = sWordStart
			goto ensured
		}

		if c != '_' {
			state = sNonDelimiter
			goto ensured
		}

		// c is _.
		switch state {
		case sWordStart:
			fallthrough
		case sNonDelimiter:
			if i < len(name)-1 {
				state = sDelimiter
				continue
			}
			state = sNonDelimiter // c is the last rune.
		}

	ensured:
		if state == sWordStart {
			b.WriteByte('_')
			b.WriteRune(unicode.ToLower(c))
		} else {
			b.WriteRune(c)
		}
	}

	return b.String()
}
