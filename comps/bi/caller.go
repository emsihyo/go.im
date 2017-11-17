package bi

import (
	"errors"
	"reflect"
)

const (
	callerHelp = "e.g. func (from interface{}), func (from interface{},req *Req), func (from interface{},req *Req) *Resp"
)

//caller caller
type caller struct {
	Func  reflect.Value
	tIn0  reflect.Type
	tIn1  reflect.Type
	tOut0 reflect.Type
}

func newCaller(v interface{}) *caller {
	c := &caller{}
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Func {
		panic(errors.New(callerHelp))
	}
	c.Func = vv
	vt := vv.Type()
	switch vt.NumIn() {
	case 2:
		c.tIn1 = vt.In(1)
		if c.tIn1.Kind() != reflect.Ptr {
			panic(errors.New(callerHelp))
		}
		fallthrough
	case 1:
		c.tIn0 = vt.In(0)
	default:
		panic(errors.New(callerHelp))
	}
	switch vt.NumOut() {
	case 1:
		c.tOut0 = vt.Out(0)
		if c.tOut0.Kind() != reflect.Ptr {
			panic(errors.New(callerHelp))
		}
	case 0:
	default:
		panic(errors.New(callerHelp))
	}
	return c
}

func (c *caller) call(from interface{}, argsProtocol Protocol, a []byte) ([]byte, error) {
	var vs []reflect.Value
	var err error
	if nil == c.tIn1 {
		vs = c.Func.Call([]reflect.Value{(reflect.ValueOf(from))})
	} else {
		in1 := reflect.New(c.tIn1.Elem()).Interface()
		if err = argsProtocol.Unmarshal(a, in1); nil != err {
			// log.Println(err)
			return nil, err
		}
		vs = c.Func.Call([]reflect.Value{reflect.ValueOf(from), reflect.ValueOf(in1)})
	}
	if 0 < len(vs) {
		var b []byte
		if b, err = argsProtocol.Marshal(vs[0].Interface()); nil != err {
			// log.Println(err)
			return nil, err
		}
		return b, nil
	}
	return nil, nil
}
