package rpcmap

import (
    "errors"
    "reflect"
    "strings"
    "unicode"
    "unicode/utf8"
)

// RPC-like calls mapper: by name, and interface{} input and output parameters which should
// match signatures of registered functions and services.
//
// It supports both services and functions, and supports multiple possible function signatures for convenience.
//
// Note that registrations (Func, ServiceDef) are not goroutine-safe, lock it yourself if you need that.
type RpcMap struct {
    fieldNameMapper func(string) string
    funcs           funcDefMap
    services        servicesMap
}

func New() *RpcMap {
    return &RpcMap{ funcs: make(funcDefMap), services: make(servicesMap), fieldNameMapper: strings.ToLower }
}

type servicesMap map[string]*ServiceDef
type funcDefMap map[string]*FuncDef

var NoMethodError = errors.New("unknown method")
var typeOfError   = reflect.TypeOf((*error)(nil)).Elem()

func methTypeCheck(t reflect.Type, extraArg int) bool {
    num_in := t.NumIn() - extraArg
    if !(t.Kind() == reflect.Func && num_in > 0 && num_in <= 2) ||
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

func (rm* RpcMap) Func(name string, f interface{}) {
    fd := makeFuncDef(f)
    rm.funcs[name] = fd
}

// Register named service. See docs of RpcMap.Func() for possible method signatures.
// Note that methods with unmatched signatures are skipped, no errors reported.
func (rm *RpcMap) NamedService(name string, rcvr interface{}) *ServiceDef {
    s := makeService(name, rcvr)
    rm.services[s.name] = s
    return s
}

// Maps all methods from passed ServiceDef instance, applying name mapper.
func (rm* RpcMap) Service(i interface{}) *ServiceDef {
    return rm.NamedService("", i)
}

// Get a function definition by registered name
func (rm* RpcMap) GetFunc(name string) *FuncDef {
    return rm.funcs[name]
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

// Get service method definition by name, using dotted notation ServiceName.Method

func (rm *RpcMap) GetServiceMethod(name string) *ServiceMethod {
    n := strings.Split(name, ".")
    if len(n) != 2 {
        return nil
    }
    sd := rm.GetService(n[0])
    if nil == sd {
        return nil
    }
    return sd.GetMethod(n[1])
}

// Call service method by name, using dotted notation ServiceName.Method
func (rm *RpcMap) CallMethod(name string, ctx interface{}, in interface{}) (interface{}, error) {
    md := rm.GetServiceMethod(name)
    if nil == md {
        return nil, NoMethodError
    }
    return md.Call(ctx, in)
}


// Sets field name mapper when registering funcs from service. Default name mapper is toLower().
// If name mapper returns empty string, method will be skipped.
func (rm *RpcMap) SetFieldNameMapper(mapper func(string) string) {
    rm.fieldNameMapper = mapper
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

