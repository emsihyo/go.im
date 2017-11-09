package bi

import "github.com/gogo/protobuf/proto"

import "fmt"

//ProtobufProtocol ProtobufProtocol
type ProtobufProtocol struct {
}

//Marshal Marshal
func (p *ProtobufProtocol) Marshal(v interface{}) ([]byte, error) {
	var err error
	defer func() {
		if e := recover(); nil != e {
			err = fmt.Errorf("proto.Marshal %v", e)
		}
	}()
	var res []byte
	m := v.(proto.Message)
	res, err = proto.Marshal(m)
	return res, err
}

//Unmarshal Unmarshal
func (p *ProtobufProtocol) Unmarshal(d []byte, v interface{}) error {
	var err error
	defer func() {
		if e := recover(); nil != e {
			err = fmt.Errorf("proto.Marshal %v", e)
		}
	}()
	m := v.(proto.Message)
	err = proto.Unmarshal(d, m)
	return err
}
