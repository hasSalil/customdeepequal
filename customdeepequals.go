package customdeepequal

import (
	"reflect"
	"unsafe"
)

// CustomDeepEquals allows registering custom equality functions for recursively
// traversed fields based on the type.
// For pointer types it deferences the pointer and then runs the comparison
// For comparison to work the input structs must be of the exact same type
type CustomDeepEquals struct {
	CustomEqualityCheckers map[reflect.Type]func(a, b unsafe.Pointer) bool
}

// NewCustomDeepEquals creates an new CustomDeepEquals
func NewCustomDeepEquals() CustomDeepEquals {
	return CustomDeepEquals{make(map[reflect.Type]func(a unsafe.Pointer, b unsafe.Pointer) bool)}
}

// RegisterEquivalenceForType registers the equals function for the given type
func (c *CustomDeepEquals) RegisterEquivalenceForType(ty reflect.Type, equals func(a, b unsafe.Pointer) bool) {
	c.CustomEqualityCheckers[ty] = equals
}

// During deepValueEqual, must keep track of checks that are
// in progress. The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

// Tests for deep equality using reflected types or custom type equivalence overrides.
// The map argument tracks comparisons that have already been seen, which allows short
// circuiting on recursive types.
func (c *CustomDeepEquals) deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}

	customEq, ok := c.CustomEqualityCheckers[v1.Type()]
	if ok {
		v1Val := unsafe.Pointer(v1.UnsafeAddr())
		v2Val := unsafe.Pointer(v2.UnsafeAddr())
		return customEq(v1Val, v2Val)
		//Can't do the it the right way below since this value might
		//be an unexported field, and go has no way to generically do
		//a cast
		//fn := reflect.ValueOf(customEq)
		//return fn.Call([]reflect.Value{v1, v2})[0].Bool()
	}
	// if depth > 10 { panic("deepValueEqual") }	// for debugging

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(t) needs to return true for at least one of the types in the cycle.
	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := unsafe.Pointer(v1.UnsafeAddr())
		addr2 := unsafe.Pointer(v2.UnsafeAddr())
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if !c.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !c.deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return c.deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return c.deepValueEqual(reflect.Indirect(v1), reflect.Indirect(v2), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !c.deepValueEqual(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !c.deepValueEqual(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	case reflect.String:
		return v1.String() == v2.String()
	default:
		// Normal equality suffices
		vb1 := (uintptr)(unsafe.Pointer(v1.UnsafeAddr()))
		vb2 := (uintptr)(unsafe.Pointer(v2.UnsafeAddr()))
		sz := v1.Type().Size()
		for i := uintptr(0); i < sz; i++ {
			v1Addr := vb1 + i
			v2Addr := vb2 + i
			b1 := *(*byte)(unsafe.Pointer(v1Addr))
			b2 := *(*byte)(unsafe.Pointer(v2Addr))
			if b1 != b2 {
				return false
			}
		}
		return true
	}
}

// DeepEqual implements a version of reflect.DeepEqual that allows override equality checking
// for registered types
func (c *CustomDeepEquals) DeepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}
	return c.deepValueEqual(v1, v2, make(map[visit]bool), 0)
}
