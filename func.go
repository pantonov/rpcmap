package rpcmap

import (
    "reflect"
)

type FuncDef struct {
    hasCtx   bool        // has context parameter
    hasRv    bool        // has context value
    argsType reflect.Type
    meth     reflect.Value
    data     map[string]interface{}
}

// Maps function to call name
// Possible function signatures:
// func(inputType) error
// func(inputType) (outputType, error)
// func(contextType, inputType) error
// func(contextType, inputType) (outputType, error)
func makeFuncDef(f interface{}) *FuncDef {
    t := reflect.TypeOf(f)
    if !methTypeCheck(t, 0) {
        panic("Invalid function signature for RpcMap.Func")
    }
    return &FuncDef{
        meth:     reflect.ValueOf(f),
        argsType: t.In(t.NumIn() - 1),
        hasCtx:   2 == t.NumIn(),
        hasRv:    2 == t.NumOut(),
        data:     make(map[string]interface{}),
    }
}

// Call a function. If function signature has no return value, returned interface is nil.
// If function does not accept context, ctx parameter will be ignored, you can pass nil.
func (fd* FuncDef) Call(ctx interface{}, in interface{}) (interface{}, error) {
    var in_args []reflect.Value
    if fd.hasCtx {
        in_args = []reflect.Value{ reflect.ValueOf(ctx), reflect.ValueOf(in) }
    } else {
        in_args = []reflect.Value{ reflect.ValueOf(in) }
    }
    rvs := fd.meth.Call(in_args)
    if fd.hasRv {
        rerr,_ := rvs[1].Interface().(error)
        return rvs[0].Interface(), rerr
    } else {
        return nil, rvs[0].Interface().(error)
    }
    return nil, nil
}

// Create input argument for function based on it's prototype. If function takes pointer, it will create
// original type.
func (fd* FuncDef) MakeInArg() interface{} {
    return reflect.New(fd.argsType.Elem()).Interface()
}

func (fd* FuncDef) Set(key string, v interface{}) {
    fd.data[key] = v
}

func (fd* FuncDef) Get(key string) interface{} {
    return fd.data[key]
}


