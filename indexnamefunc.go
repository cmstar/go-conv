package conv

import (
	"strings"
	"unicode"
)

// Provides some built-in IndexNameFunc for different situations.

// IndexNameFunc defines a function used to match names when converting from map to struct or from struct to struct.
//
// If the given name can match, the function returns the value from the source map with @ok=true;
// otherwise returns (nil, false) .
// If it returns OK, the value from the source map will be converted into the destination struct
// using Conv.ConvertType() .
//
// When converting a map to a struct, each field name of the struct will be indexed using this function.
// When converting a struct to another, field names and values from the source struct will be put into a map,
// then each field name of the destination struct will be indexed with the map.
//
type IndexNameFunc func(m map[string]interface{}, key string) (value interface{}, ok bool)

// CaseInsensitiveIndexName is IndexNameFunc, it indexes a map and compares the keys case-insensitively.
// The function use strings.EqualFold() to compare keys, returns on the first key which strings.EqualFold() is true.
func CaseInsensitiveIndexName(m map[string]interface{}, key string) (value interface{}, ok bool) {
	return iterateAllKeys(m, key, strings.EqualFold)
}

func iterateAllKeys(m map[string]interface{}, key string, comp func(x, y string) bool) (interface{}, bool) {
	// No build-in method to index a map case-insensitively, we just iterate all keys.
	for k, v := range m {
		if comp(key, k) {
			return v, true
		}
	}

	return nil, false
}

// CamelSnakeCaseIndexName is IndexNameFunc, it can match names in camel-case - like 'lowerCaseCamel' or 'UpperCaseCamel'
// - or snake-case - like 'snake_case' or 'Snake_Case' (this style is sometimes called train-case).
// Specially, when comparing two names, if anyone contains a space rune, the names must be equal strictly.
//
// This function use strings.EqualFold() to compare the first rune of each word,
// and use equal operator (==) for other runes, ignoring the first underscore (_) before each word.
//
// These names are equal: oneTwoThree, OneTwoThree, one_two_three, One_Two_Three, one_TwoThree
//
// These names are not equal: _oneTwoThree, OneTwoThree, onetwoThree, OneTwoThree_, one__two_three
//
// Mostly this function can be used to match field names from different platform, e.g.,
// 'lowerCaseCamel' from Javascript, 'UpperCaseCamel' from Go, 'snake_case' from Mysql database.
//
func CamelSnakeCaseIndexName(m map[string]interface{}, key string) (value interface{}, ok bool) {
	return iterateAllKeys(m, key, camelSnakeCaseCompare)
}

func camelSnakeCaseCompare(sx, sy string) bool {
	x, y := []rune(sx), []rune(sy)
	lenX, lenY := len(x), len(y)
	if lenX == 0 && lenY == 0 {
		return true
	}
	if lenX == 0 || lenY == 0 {
		return false
	}

	iterX := camelSnakeNameIter{s: []rune(sx)}
	iterY := camelSnakeNameIter{s: []rune(sy)}
	for {
		iterX.next()
		iterY.next()

		if iterX.idx == -1 && iterY.idx == -1 {
			// The end, all runes ahead are equal.
			return true
		}

		if iterX.idx == -1 || iterY.idx == -1 {
			return false
		}

		// If a name contains any space rune, the two name must be equal strictly.
		if unicode.IsSpace(iterX.Current) || unicode.IsSpace(iterY.Current) {
			return sx == sy
		}

		if iterX.IsWordStart && iterY.IsWordStart {
			if !strings.EqualFold(string(iterX.Current), string(iterY.Current)) {
				return false
			}

			continue
		}

		if iterX.IsWordStart || iterY.IsWordStart {
			return false
		}

		if iterX.Current != iterY.Current {
			return false
		}
	}
}

type camelSnakeNameIter struct {
	s           []rune // The whole string.
	idx         int    // The next index use by next(), increased after next() is called, -1 if next() at the end of s.
	IsWordStart bool   // If the current rune is a start of a word.
	Current     rune   // The current rune during the iteration.
}

func (iter *camelSnakeNameIter) next() {
	if iter.idx >= len(iter.s) {
		iter.idx = -1
		iter.IsWordStart = false
		iter.Current = 0
		return
	}

	// IsWordStart if any of:
	// 1. The first rune.
	// 2. An uppercase rune after a lowercase rune.
	// 3. A rune after a single underscore, and the underscore is not the first rune.

	// Case 1 & 2.
	cur := iter.s[iter.idx]
	if iter.idx == 0 {
		iter.IsWordStart = true
		iter.Current = cur
		iter.idx++
		return
	}

	// Case 2.
	prev := iter.s[iter.idx-1]
	if unicode.IsUpper(cur) && unicode.IsLower(prev) {
		iter.IsWordStart = true
		iter.Current = cur
		iter.idx++
		return
	}

	// Case 3.
	if cur == '_' && prev != '_' && iter.idx != len(iter.s)-1 {
		// Skip the current rune which is the delimiter underscore of snake-case style.
		iter.IsWordStart = true
		iter.Current = iter.s[iter.idx+1]
		iter.idx += 2
		return
	}

	iter.IsWordStart = false
	iter.Current = cur
	iter.idx++
}
