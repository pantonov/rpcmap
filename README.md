[![Build Status](https://travis-ci.org/pantonov/rpcmap.svg)](https://travis-ci.org/pantonov/rpcmap) 
[![GoDoc](https://godoc.org/github.com/pantonov/rpcmap?status.svg)](https://godoc.org/github.com/pantonov/rpcmap)

## RPC-like function and service mapper

* Optimized for use in RPC scenarios
* Registration of service methods and plain functions
* Call functions by name, call service methods by scoped name (Service.Method) or by method name only
* Supports several function/method signatures (see below)
* Method name mapper, per-method meta-data
* stdlib-only dependencies

## Function/method signatures
```go
func() error
func() (outputType, error)
func(inputType) error
func(inputType) (outputType, error)
func(contextType, inputType) error
func(contextType, inputType) (outputType, error)
```
When registering a service, methods with signatures which do not match to any of the above are skipped.

## Examples

### Service registration

```go
type MyService struct {}

type testInput struct {
    Foo string
}

function testOutput struct {
    Greeting string
}

// rpcmap will find this method on registration
func (s *MyService) Hello(i *testInput) (*testOutput, error) {
    return &testOutput{ Greeting: fmt.Sprintf("Hi,%s!", i.Foo) }, nil
}
    
mapper := rpcmap.New()

mapper.Service(&MyService{}) // register service

mapper.Func("f3", func(i int) (string, error) { // register function
        return fmt.Sprintf("Hi,%d", i), nil
})

// Call service method with dotted notation. Note that the method name is lowercase (uses default mapper)
// To use method names only, register service with DefaultService()
rv1, err := mapper.CallMethod("S.hello", nil, &testInput{ Foo: "zz"})
fmt.Printf("rv.greeting = %s", rv.(*testOutput).Greeting)

// typical RPC case (error handling omitted for simplicity)
meth := mapper.GetCallable("S.hello")
arg := meth.MakeArg() // if function has no args, returns Ptr to empty struct instance
json.Unmarshal(data, arg)
rv, err := meth->Call(myContext, arg)
if nil != err {
    result,_ = json.Marshal(rv)
}

```

For more information, see [Documentation](https://godoc.org/github.com/pantonov/rpcmap)




### License
[MIT](https://opensource.org/licenses/MIT).
