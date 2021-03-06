package rpcmap

import (
    "reflect"
)

// Function definition object
type FuncDef struct {
    argno    int
    hasRv    bool
    argsType reflect.Type
    meth     reflect.Value
    data     map[string]interface{}
}

func makeFuncDef(f interface{}) *FuncDef {
    t := reflect.TypeOf(f)
    if !methTypeCheck(t, 0) {
        panic("Invalid function signature for RpcMap.Func")
    }
    fd := &FuncDef{
        meth:     reflect.ValueOf(f),
        argno:    t.NumIn(),
        hasRv:    2 == t.NumOut(),
        data:     make(map[string]interface{}),
    }
    if t.NumIn() == 0 {
        fd.argsType = typeOfEmptyStruct
    } else {
        fd.argsType = t.In(t.NumIn() - 1)
    }
    return fd
}

// Call a function. If function signature has no return value, returned interface is nil.
// If function does not accept context, ctx parameter will be ignored, you can pass nil.
func (fd* FuncDef) Call(ctx interface{}, in interface{}) (interface{}, error) {
    var in_args []reflect.Value
    switch fd.argno {
    case 0:
        in_args = []reflect.Value{}
    case 1:
        in_args = []reflect.Value{ reflect.ValueOf(in) }
    case 2:
        in_args = []reflect.Value{ reflect.ValueOf(ctx), reflect.ValueOf(in) }
    }
    rvs := fd.meth.Call(in_args)
    if fd.hasRv {
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
func (fd* FuncDef) MakeArg() interface{} {
    return makeArg(fd.argsType)
}

// Returns number of function arguments
func (fd* FuncDef) InArgs() int {
    return fd.argno
}

// Returns true if functions has a result
func (fd* FuncDef) HasOutArg() bool {
    return fd.hasRv
}

// Set function meta-data value
func (fd* FuncDef) Set(key string, v interface{}) {
    fd.data[key] = v
}

// Get function meta-data value
func (fd* FuncDef) Get(key string) interface{} {
    return fd.data[key]
}


