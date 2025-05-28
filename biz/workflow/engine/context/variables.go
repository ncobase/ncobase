package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	// ErrKeyNotFound is returned when a key doesn't exist
	ErrKeyNotFound = errors.New("key not found")
	// ErrNilValue is returned when attempting to set a nil value
	ErrNilValue = errors.New("nil value not allowed")
)

// Variables manages variables in a thread-safe way
type Variables struct {
	mu    sync.RWMutex
	data  map[string]any
	dirty bool
}

// NewVariables creates a new variables manager with optional initial capacity
//
// Usage:
//
//	// Create a new Variables instance
//	vars := NewVariables()
//
//	// Create with initial capacity
//	vars := NewVariables(10)
//
//	// Basic operations
//	vars.Set("string_key", "value")
//	vars.Set("int_key", 42)
//	vars.Set("map_key", map[string]any{"nested": "value"})
//
//	// Get values
//	val, err := vars.Get("string_key")
//	defaultVal := vars.GetDefault("missing_key", "default")
//
//	// Check and delete
//	if vars.HasKey("int_key") {
//	    vars.Delete("int_key")
//	}
//
//	// Merge variables
//	other := NewVariables()
//	other.Set("new_key", "new_value")
//	if err := vars.Merge(other); err != nil {
//	    // Handle error
//	}
//
//	// Clone and modify
//	clone, err := vars.Clone()
//	if err != nil {
//	    // Handle error
//	}
func NewVariables(capacity ...int) *Variables {
	c := 0
	if len(capacity) > 0 {
		c = capacity[0]
	}
	return &Variables{
		data: make(map[string]any, c),
	}
}

// Get retrieves a variable value by key
func (v *Variables) Get(key string) (any, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	val, ok := v.data[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	return val, nil
}

// GetDefault retrieves a value with a default fallback
func (v *Variables) GetDefault(key string, defaultVal any) any {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if val, ok := v.data[key]; ok {
		return val
	}
	return defaultVal
}

// MustGet gets a value or panics if not found
func (v *Variables) MustGet(key string) any {
	val, err := v.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Set stores a variable value
func (v *Variables) Set(key string, value any) error {
	if value == nil {
		return ErrNilValue
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.data[key] = value
	v.dirty = true
	return nil
}

// SetIfNotExists sets a value only if the key is not present
func (v *Variables) SetIfNotExists(key string, value any) (bool, error) {
	if value == nil {
		return false, ErrNilValue
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.data[key]; exists {
		return false, nil
	}

	v.data[key] = value
	v.dirty = true
	return true, nil
}

// Delete removes variables and returns existence status
func (v *Variables) Delete(keys ...string) []bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	results := make([]bool, len(keys))
	for i, key := range keys {
		_, exists := v.data[key]
		if exists {
			delete(v.data, key)
			v.dirty = true
			results[i] = true
		}
	}
	return results
}

// Clear removes all variables and returns count
func (v *Variables) Clear() int {
	v.mu.Lock()
	defer v.mu.Unlock()

	count := len(v.data)
	v.data = make(map[string]any)
	v.dirty = count > 0
	return count
}

// Len returns the number of variables
func (v *Variables) Len() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.data)
}

// Keys returns all variable keys
func (v *Variables) Keys() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	keys := make([]string, 0, len(v.data))
	for k := range v.data {
		keys = append(keys, k)
	}
	return keys
}

// HasKey checks if a key exists
func (v *Variables) HasKey(key string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	_, ok := v.data[key]
	return ok
}

// IsDirty checks modification status
func (v *Variables) IsDirty() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.dirty
}

// ClearDirty resets modification status
func (v *Variables) ClearDirty() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.dirty = false
}

// mergeOptions defines merge behavior
type mergeOptions struct {
	override bool
}

// MergeOption is a function that configures merge options
type MergeOption func(*mergeOptions)

// WithOverride configures whether to override existing keys
func WithOverride(override bool) MergeOption {
	return func(o *mergeOptions) {
		o.override = override
	}
}

// Merge combines variables from another instance
func (v *Variables) Merge(other *Variables, opts ...MergeOption) error {
	if other == nil {
		return nil
	}

	options := mergeOptions{
		override: true, // default to override
	}
	for _, opt := range opts {
		opt(&options)
	}

	other.mu.RLock()
	defer other.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	for k, val := range other.data {
		if !options.override {
			if _, exists := v.data[k]; exists {
				continue
			}
		}

		copyVal, err := deepCopy(val)
		if err != nil {
			return fmt.Errorf("failed to copy value for key %s: %v", k, err)
		}
		v.data[k] = copyVal
	}

	v.dirty = true
	return nil
}

// MergeJSON merges variables from JSON data
func (v *Variables) MergeJSON(jsonData any) error {
	if jsonData == nil {
		return nil
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON data: %v", err)
	}

	tempVars := NewVariables()
	if err := tempVars.UnmarshalJSON(data); err != nil {
		return err
	}

	return v.Merge(tempVars)
}

// Clone creates a deep copy of variables
func (v *Variables) Clone() (*Variables, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	clone := NewVariables(len(v.data))
	for k, val := range v.data {
		copyVal, err := deepCopy(val)
		if err != nil {
			return nil, fmt.Errorf("failed to copy value for key %s: %v", k, err)
		}
		clone.data[k] = copyVal
	}
	clone.dirty = v.dirty
	return clone, nil
}

// MarshalJSON implements json.Marshaler
func (v *Variables) MarshalJSON() ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return json.Marshal(v.data)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *Variables) UnmarshalJSON(data []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.data == nil {
		v.data = make(map[string]any)
	}
	return json.Unmarshal(data, &v.data)
}

// deepCopy creates a deep copy of any value
func deepCopy(src any) (any, error) {
	if src == nil {
		return nil, nil
	}

	// Handle primitive types directly
	switch v := src.(type) {
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64:
		return v, nil
	case []byte:
		result := make([]byte, len(v))
		copy(result, v)
		return result, nil
	case time.Time:
		return v, nil
	}

	// Handle composite types
	kind := reflect.TypeOf(src).Kind()
	switch kind {
	case reflect.Invalid:
		return nil, fmt.Errorf("cannot deep copy invalid kind")
	case reflect.Slice:
		return deepCopySlice(src)
	case reflect.Map:
		return deepCopyMap(src)
	default:
		// Fallback to JSON for complex types
		return deepCopyJSON(src)
	}
}

// deepCopySlice handles slice deep copying
func deepCopySlice(src any) (any, error) {
	srcVal := reflect.ValueOf(src)
	newSlice := reflect.MakeSlice(srcVal.Type(), srcVal.Len(), srcVal.Cap())
	for i := 0; i < srcVal.Len(); i++ {
		itemCopy, err := deepCopy(srcVal.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		newSlice.Index(i).Set(reflect.ValueOf(itemCopy))
	}
	return newSlice.Interface(), nil
}

// deepCopyMap handles map deep copying
func deepCopyMap(src any) (any, error) {
	srcVal := reflect.ValueOf(src)
	newMap := reflect.MakeMap(srcVal.Type())
	iter := srcVal.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		copyVal, err := deepCopy(v.Interface())
		if err != nil {
			return nil, err
		}
		newMap.SetMapIndex(k, reflect.ValueOf(copyVal))
	}
	return newMap.Interface(), nil
}

// deepCopyJSON handles complex type copying via JSON
func deepCopyJSON(src any) (any, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %v", err)
	}

	var dest any
	if err := json.Unmarshal(data, &dest); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}

	return dest, nil
}
