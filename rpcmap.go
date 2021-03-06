// RPC-like calls mapper: by name, and interface{} input and output parameters which should
// match signatures of registered functions and services.
//
// It supports both services and functions, and allows several possible function signatures for convenience.
//
// Note that registrations (Func, ServiceDef) are not goroutine-safe, lock it yourself if you need that.
//
// See README.md and rpcmap_test.go for usage examples.
//
package rpcmap

import (
    "errors"
    "reflect"
    "strings"
    "unicode"
    "unicode/utf8"
)

// Function and method mapper instance
type RpcMap struct {
    fieldNameMapper func(string) string
    funcs           funcDefMap
    services        servicesMap
    defaultService  *ServiceDef
}

// Callable object (function or method)
type Callable interface {
    // Call method or function with context (may be nil) and input argument inArg
    Call(ctx interface{}, inArg interface{}) (interface{}, error)

    // create instance of input argument of a method. If method has no arguments, this method
    // will return instance of NoArg (synonym of struct{}), to keep *.Unmarshal happy.
    MakeArg() interface{}

    // Number of input arguments (0 if no args, 1 if input arg only, 2 if input arg and context)
    InArgs() int

    // Has result value, not only error
    HasOutArg() bool

    // Set and get arbitrary function/method definition meta-data (e.g. privilege level for run-time checking)
    Set(key string, value interface{})
    Get(key string) interface{}
}

func New() *RpcMap {
    return &RpcMap{ funcs: make(funcDefMap), services: make(servicesMap), fieldNameMapper: strings.ToLower }
}

// Hint: use `_*rpcmap.NoArg` as a second parameter in methods which take context but do not take input argument
// e.g. `func(ctx myContext, _*rpcmap.NoArg)` error
type NoArg = struct{}

type servicesMap map[string]*ServiceDef
type funcDefMap map[string]*FuncDef

var NoMethodError = errors.New("unknown method")
var typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
var typeOfEmptyStruct = reflect.TypeOf(struct{}{})

func methTypeCheck(t reflect.Type, extraArg int) bool {
    num_in := t.NumIn() - extraArg
    if !(t.Kind() == reflect.Func && num_in >= 0 && num_in <= 2) ||
        !(t.NumOut() == 1 || t.NumOut() == 2) {
            return false
    }
    switch t.NumOut() {
    case 1:
        if !(t.Out(0).Implements(typeOfError)) {
            return false
        }
    case 2:
        if !(t.Out(1).Implements(typeOfError)) {
            return false
        }
    default:
        return false
    }
    return true
}

func makeArg(t reflect.Type) interface{} {
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return reflect.New(t).Interface()
}

// Register function with name. Possible function signatures:
// func() error
// func() (outputType, error)
// func(inputType) error
// func(inputType) (outputType, error)
// func(contextType, inputType) error
// func(contextType, inputType) (outputType, error)
func (rm* RpcMap) Func(name string, f interface{}) Callable {
    fd := makeFuncDef(f)
    rm.funcs[name] = fd
    return fd
}

// Register service with explicit name (overriding default, which is name of service type).
// See docs of RpcMap.Func() for possible method signatures.
// Note that methods with not allowed signatures are skipped, no errors reported.
func (rm *RpcMap) NamedService(name string, rcvr interface{}) *ServiceDef {
    s := makeService(rm.fieldNameMapper, name, rcvr)
    rm.services[s.name] = s
    return s
}

// Register service with name equal to rcvr type name.
// Note that methods with not allowed signatures are skipped, no errors reported.
func (rm* RpcMap) Service(rcvr interface{}) *ServiceDef {
    return rm.NamedService("", rcvr)
}

// Registers service and makes it default, so its method can be called with method name only.
// Note that methods with not allowed signatures are skipped, no errors reported.
// Only single default service can be registered
func (rm* RpcMap) DefaultService(rcvr interface{}) *ServiceDef {
    rm.defaultService = rm.Service(rcvr)
    return rm.defaultService
}

// Get a function callable object by registered name
func (rm* RpcMap) GetFunc(name string) Callable {
    f := rm.funcs[name]
    if nil != f {
        return f
    }
    return nil
}

// Get a function definition by registered name
func (rm* RpcMap) GetService(name string) *ServiceDef {
    return rm.services[name]
}

// Call registered function by name.
func (rm *RpcMap) CallFunc(name string, ctx interface{}, in interface{}) (interface{}, error) {
    md := rm.GetFunc(name)
    if nil == md {
        return nil, NoMethodError
    }
    return md.Call(ctx, in)
}

// Get service method callable object by name, using dotted notation ServiceName.Method,
// or just method name, if service was registered wuth DefaultService().
func (rm *RpcMap) GetServiceMethod(name string) Callable {
    n := strings.Split(name, ".")
    switch len(n) {
    case 2:
        sd := rm.GetService(n[0])
        if nil == sd {
            return nil
        }
        return sd.GetMethod(n[1])
    case 1:
        if nil == rm.defaultService {
            return nil
        }
        return rm.defaultService.GetMethod(n[0])
    }
    return nil
}

// Returns list of registered services
func (rm* RpcMap) ListServices() []*ServiceDef {
    sl := make([]*ServiceDef, 0, 16)
    for _, s := range rm.services {
        sl = append(sl, s)
    }
    return sl
}

// Returns list of registered functions
func (rm *RpcMap) ListFunctions() []*FuncDef {
    fl := make([]*FuncDef, 0, 16)
    for _, f := range rm.funcs {
        fl = append(fl, f)
    }
    return fl
}

// Returns registered function callable by name, or method callable from default service
func (rm* RpcMap) GetCallable(name string) Callable {
    if c := rm.GetFunc(name); nil != c {
        return c
    }
    if nil == rm.defaultService {
        return nil
    }
    return rm.defaultService.GetMethod(name)
}

// Call service method by name, using dotted notation ServiceName.Method
func (rm *RpcMap) CallMethod(name string, ctx interface{}, in interface{}) (interface{}, error) {
    md := rm.GetServiceMethod(name)
    if nil == md {
        return nil, NoMethodError
    }
    return md.Call(ctx, in)
}

// Sets field name mapper when registering methods for service. Default name mapper is toLower().
// If name mapper returns empty string, method will be skipped.
func (rm *RpcMap) SetFieldNameMapper(mapper func(string) string) {
    rm.fieldNameMapper = mapper
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

