# A Customizable DeepEqual implementation for golang
A re-implementation of reflect.DeepEqual that allows registering of custom deep equal function for specified types. The custom deep equals work recursively when comparing structs


## Caution
This library's DeepEqual method is reflection-based, and not efficient. It is therefore not suitable for use in production, and is intended for aid in unit testing and debugging.


## Sample Usage
```
customDeep := NewCustomDeepEquals()
customDeep.RegisterEquivalenceForType(reflect.TypeOf(time.Now()), func(a, b unsafe.Pointer) bool {
    // Make sure to cast to pointer of the type registered in the first param
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
if customDeep.DeepEqual(&ctype{c: 1, d: &dtype{d: &two, t: &t1}, e: &[]byte{0, 1}, f: &str, t: &t1}, &ctype{c: 1, d: &dtype{d: &two2, t: &t2}, e: &[]byte{0, 1}, f: &str2, t: &t2}) {
    // do stuff...
}
```
