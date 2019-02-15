package rpcmap

import (
    "encoding/json"
    "fmt"
    "testing"
)

type testCtx struct {
    s  string
}

type testInput struct {
    A string
    B int
}

type testResult struct {
    o   string
}

func TestF(t *testing.T) {
    r := New()
    r.Func("f1", func(i testInput) error {
        return fmt.Errorf("Hi,%s/%d", i.A, i.B)
    })
    r.Func("f2", func(ctx* testCtx, i testInput) error {
        return fmt.Errorf("Hi,%s/%d/%s", i.A, i.B, ctx.s)
    })
    r.Func("f3", func(i int) error {
        return fmt.Errorf("Hi,%d", i)
    })
    r.Func("f4", func() (string, error) {
        return "blah", nil
    })
    _, err1 := r.CallFunc("f1", nil, testInput{A: "xx", B:5 })
    _, err2 := r.CallFunc("f2", &testCtx{s: "T"}, testInput{A: "xx", B:6 })
    _, err3 := r.CallFunc("f3",nil, 3)
    if r4, _ := r.CallFunc("f4", nil, nil); r4 != "blah" {
        t.FailNow()
    }
    if err1.Error() != "Hi,xx/5" || err2.Error() != "Hi,xx/6/T" || err3.Error() != "Hi,3" {
        t.FailNow()
    }
    r.Func("f3", func(i int) (string, error) {
        return fmt.Sprintf("Hi,%d", i), nil
    })
    r.Func("f4", func(ctx testCtx, i int) (string, error) {
        return fmt.Sprintf("Hi,%s/%d", ctx.s, i), nil
    })
    rv1, _ := r.CallFunc("f3", nil, 7)
    rv2, _ := r.CallFunc("f4", testCtx{s: "Ho"}, 8)
    if rv1 != "Hi,7" || rv2 != "Hi,Ho/8" {
        t.FailNow()
    }
}

type S struct {}

func (s *S) Meth1(i *testInput) error {
    return fmt.Errorf("Hi,%s/%d", i.A, i.B)
}

func (s *S) Meth2(ctx* testCtx, i *testInput) error {
    return fmt.Errorf("Hi,%s/%d/%s", i.A, i.B, ctx.s)
}

func (s *S) Meth3(ctx* testCtx, i *testInput) (*testResult, error) {
    out := fmt.Sprintf("Hi,%s/%d/%s", i.A, i.B, ctx.s)
    return &testResult{ o: out }, nil
}

func (s *S) Meth4() (*testResult, error) {
    return &testResult{ o: "hi" }, nil
}

func TestS(t *testing.T) {
    r := New()
    svc := S{}
    r.Service(&svc)
    if _, err := r.CallMethod("S.meth1", nil, &testInput{A: "zz", B:3 }); err.Error() != "Hi,zz/3" {
        t.Fatalf("S.meth1")
    }
    sd := r.GetService("S")
    if nil == sd {
        t.Fatalf("Service not found")
    }
    md := sd.GetMethod("meth2")
    if nil == md {
        t.Fatalf("Method not found")
    }
    if _, err1 := md.Call(&testCtx{s:"ku"}, &testInput{A: "yy", B:2 }); err1.Error() != "Hi,yy/2/ku" {
        t.Fatalf("md.Call")
    }
    arg := md.MakeInArg()
    if err := json.Unmarshal([]byte(`{"A":"xx","B":9}`), arg); nil != err {
        t.Fatal(err)
    }
    if _, err1 := md.Call(&testCtx{s:"uu"}, &testInput{A: "kk", B:6 }); err1.Error() != "Hi,kk/6/uu" {
        t.Fatal(err1)
    }
    res, _ := r.CallMethod("S.meth3", &testCtx{s:"cc"}, &testInput{A: "tt", B:77 })
    if res.(*testResult).o != "Hi,tt/77/cc" {
        t.Fatalf("S.Meth3 with testResult")
    }
    res4, _ := r.CallMethod("S.meth4", nil, nil)
    if res4.(*testResult).o != "hi" {
        t.Fatalf("S.Meth4 fail")
    }
}



