package rpcmap

import (
    "fmt"
    "reflect"
)

// Service methods mapper object
type ServiceDef struct {
    name     string
    rcvr     reflect.Value         // receiver of funcs for the service
    methods  map[string]*MethodDef // registered funcs
}

// Method definition object
type MethodDef struct {
    hasRv       bool        // has context value
    argno       int

    name      string         // mapped name
    origname  string         // original name name
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
        methods:  make(map[string]*MethodDef),
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
        name := fm(method.Name)
        s.methods[name] = &MethodDef{
            origname:  method.Name,
            name:      name,
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

// Iterate over mapped method definitions and remove these for which filter_func returns false.
// This can be used for partial service mapping.
func (sd* ServiceDef) Filter(filter_func func(md*MethodDef) bool) {
    for _, m := range sd.methods {
        if !filter_func(m) {
            delete(sd.methods, m.name)
        }
    }
}

// Get a service method definition by name
func (sd *ServiceDef) GetMethod(name string) Callable {
    m := sd.methods[name]
    if nil != m {
        return m
    }
    return nil
}

// Call a method f function signature has no return value, returned interface is nil.
// If method does not accept context, ctx parameter will be ignored, you can pass nil.
func (ms *MethodDef) Call(ctx interface{}, in interface{}) (interface{}, error) {
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

// Original method name, without name mapping applied
func (ms* MethodDef) Name() string {
    return ms.origname
}

// Create input argument for function based on it's prototype. If function takes pointer, it will create
// original type.
func (ms *MethodDef) MakeArg() interface{} {
    return makeArg(ms.argsType)
}

// Returns number of method arguments (not counting receiver).
func (ms* MethodDef) InArgs() int {
    return ms.argno - 1
}

// Returns true if method has a result
func (ms* MethodDef) HasOutArg() bool {
    return ms.hasRv
}

// Returns list of registered methods
func (sd *ServiceDef) ListMethods() []*MethodDef {
    ml := make([]*MethodDef, 0, 16)
    for _, s := range sd.methods {
        ml = append(ml, s)
    }
    return ml
}

// Set method meta-data value
func (s*MethodDef) Set(key string, v interface{}) {
    s.data[key] = v
}

// Get method meta-data value
func (s*MethodDef) Get(key string) interface{} {
    return s.data[key]
}