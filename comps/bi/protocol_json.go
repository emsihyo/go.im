package bi

import (
	"encoding/json"
	"fmt"
)

//JSONProtocol JSONProtocol
type JSONProtocol struct {
}

//Marshal Marshal
func (p *JSONProtocol) Marshal(v interface{}) ([]byte, error) {
	var err error
	defer func() {
		if e := recover(); nil != e {
			err = fmt.Errorf("json.Marshal %v", e)
		}
	}()
	var res []byte
	res, err = json.Marshal(v)
	return res, err
}

//Unmarshal Unmarshal
func (p *JSONProtocol) Unmarshal(d []byte, v interface{}) error {
	var err error
	defer func() {
		if e := recover(); nil != e {
			err = fmt.Errorf("json.Marshal %v", e)
		}
	}()
	err = json.Unmarshal(d, v)
	return err
}
