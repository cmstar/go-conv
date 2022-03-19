//go:build debug

package conv

import "sync"

/*
Mock sync.Map. It is hard to use in debugging, so we make this one.
Add build -tags=debug to enable this.
*/

// syncMap is like sync.Map.
// The zero value is ready for use. It must not be copied after first use.
//
// The thread-safe implementation is simply based on a sync.RWMutex. It doesn't
// support some features such as netsted Range() - which will cause a dead lock.
//
type syncMap struct {
	mu   sync.RWMutex
	data map[any]any
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (s *syncMap) LoadOrStore(key, value any) (actual any, loaded bool) {
	s.mu.RLock()
	if s.data == nil {
		s.data = make(map[any]any)
	} else {
		actual, loaded = s.data[key]
	}
	s.mu.RUnlock()

	if loaded {
		return
	}

	s.mu.Lock()
	if actual, loaded = s.data[key]; !loaded {
		s.data[key] = value
		actual = value
	}
	s.mu.Unlock()
	return
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (s *syncMap) Load(key any) (value any, ok bool) {
	s.mu.RLock()
	if s.data != nil {
		value, ok = s.data[key]
	}
	s.mu.RUnlock()
	return
}

// Store sets the value for a key.
func (s *syncMap) Store(key, value any) {
	s.mu.Lock()
	if s.data == nil {
		s.data = make(map[any]any)
	}
	s.data[key] = value
	s.mu.Unlock()
}

// Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.
func (s *syncMap) Range(f func(key, value any) bool) {
	s.mu.RLock()
	for k, v := range s.data {
		if !f(k, v) {
			break
		}
	}
	s.mu.RUnlock()
}

// Delete deletes the value for a key.
func (s *syncMap) Delete(key any) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (s *syncMap) LoadAndDelete(key any) (value any, loaded bool) {
	s.mu.Lock()
	if s.data != nil {
		value, loaded = s.data[key]
		if loaded {
			delete(s.data, key)
		}
	}
	s.mu.Unlock()
	return
}
