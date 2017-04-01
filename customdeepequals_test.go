package customdeepequal

import (
	"reflect"
	"testing"
	"time"
	"unsafe"
)

type ctype struct {
	c int
	d *dtype
	e *[]byte
	f *string
	t *time.Time
}

type dtype struct {
	d *int
	t *time.Time
}

func TestAll(t *testing.T) {
	customDeep := CustomDeepEquals{make(map[reflect.Type]func(a unsafe.Pointer, b unsafe.Pointer) bool)}
	now := time.Now()
	customDeep.RegisterEquivalenceForType(reflect.TypeOf(now), func(a, b unsafe.Pointer) bool {
		// Ugly code everyone will be forced to write
		aT := (*time.Time)(a)
		bT := (*time.Time)(b)
		return aT.Unix() == bT.Unix()
	})
	two := 2
	two2 := 2
	str := "sdgv"
	str2 := "sdgv"
	t1 := time.Date(2017, 4, 1, 0, 0, 0, 24534, time.UTC)
	t2 := time.Date(2017, 4, 1, 0, 0, 0, 33454, time.UTC)
	if !customDeep.DeepEqual(&ctype{c: 1, d: &dtype{d: &two, t: &t1}, e: &[]byte{0, 1}, f: &str, t: &t1}, &ctype{c: 1, d: &dtype{d: &two2, t: &t2}, e: &[]byte{0, 1}, f: &str2, t: &t2}) {
		t.Fatalf("The structs were not deep equal")
	}
	t.Logf("The structs were deep equal")
}
