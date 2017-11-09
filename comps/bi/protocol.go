package bi

//Protocol Protocol
type Protocol interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(d []byte, v interface{}) error
}
