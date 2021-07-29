package conv

import (
	"strings"
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
// When converting a struct to another, field names and values from the souce struct will be put into a map,
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
