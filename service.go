package rpcmap

import (
    "fmt"
    "reflect"
)

// Defines a service mapping
type ServiceDef struct {
    name     string
    rcvr     reflect.Value             // receiver of funcs for the service
    methods  map[string]*ServiceMethod // registered funcs
}

type ServiceMethod struct {
    hasRv       bool        // has context value
    argno       int

    method    reflect.Method // receiver method
    argsType  reflect.Type   // type of the request argument
    rcvr      *reflect.Value
    data     map[string]interface{}
}


// Register named service. See docs of RpcMap.Func() for possible method signatures.
// Note that funcs with not matched signatures are skipped, no errors reported.
func makeService(fm func(string) string, name string, rcvr interface{}) (s *ServiceDef) {
    s = &ServiceDef{
        name:     name,
        rcvr:     reflect.ValueOf(rcvr),
        methods:  make(map[string]*ServiceMethod),
    }
    rcvrType := reflect.TypeOf(rcvr)

    if name == "" {
        s.name = reflect.Indirect(s.rcvr).Type().Name()
        if !isExported(s.name) {
            panic(fmt.Sprintf("rpc: type %q is not exported", s.name))
        }
    }
    if s.name == "" {
        panic(fmt.Sprintf("rpc: no service name for type %s", rcvrType))
    }
    for i := 0; i < rcvrType.NumMethod(); i++ {
        method := rcvrType.Method(i)
        mtype := method.Type
        if method.PkgPath != "" {
            continue
        }
        if !methTypeCheck(mtype, 1) {
            continue
        }
        s.methods[fm(method.Name)] = &ServiceMethod{
            method:    method,
            argsType:  mtype.In(mtype.NumIn() - 1),
            rcvr:      &s.rcvr,
            argno:      mtype.NumIn(),
            hasRv:      2 == mtype.NumOut(),
            data:     make(map[string]interface{}),
        }
    }
    return
}

func (s *ServiceDef) GetMethod(name string) Callable {
    m := s.methods[name]
    if nil != m {
        return m
    }
    return nil
}

// Call a method f function signature has no return value, returned interface is nil.
// If method does not accept context, ctx parameter will be ignored, you can pass nil.
func (ms *ServiceMethod) Call(ctx interface{}, in interface{}) (interface{}, error) {
    var in_args []reflect.Value
    switch ms.argno {
    case 3:
        in_args = []reflect.Value{ *ms.rcvr, reflect.ValueOf(ctx), reflect.ValueOf(in) }
    case 2:
        in_args = []reflect.Value{ *ms.rcvr, reflect.ValueOf(in) }
    case 1:
        in_args = []reflect.Value{ *ms.rcvr }
    }
    rvs := ms.method.Func.Call(in_args)
    if ms.hasRv {
        rerr,_ := rvs[1].Interface().(error)
        return rvs[0].Interface(), rerr
    } else {
        rerr,_ := rvs[0].Interface().(error)
        return nil, rerr
    }
    return nil, nil
}

// Create input argument for function based on it's prototype. If function takes pointer, it will create
// original type.
func (s *ServiceMethod) MakeInArg() interface{} {
    return makeArg(s.argsType)
}

func (s* ServiceMethod) Set(key string, v interface{}) {
    s.data[key] = v
}

func (s* ServiceMethod) Get(key string) interface{} {
    return s.data[key]
}