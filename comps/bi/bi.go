package bi

//BI BI
type BI struct {
	callers map[string]*caller
}

//NewBI NewBI
func NewBI() *BI {
	return &BI{callers: map[string]*caller{}}
}

//On On
func (bi *BI) On(m string, f interface{}) {
	bi.callers[m] = newCaller(f)
}

//Handle Handle
func (bi *BI) Handle(sess Session) {
	sess.GetSessionImpl().handle(bi, sess)
}

func (bi *BI) onEmit(from interface{}, m string, protocol Protocol, a []byte) {
	if caller, ok := bi.callers[m]; ok {
		caller.call(from, protocol, a)
	}
}

func (bi *BI) onRequest(from interface{}, m string, protocol Protocol, a []byte) ([]byte, error) {
	if caller, ok := bi.callers[m]; ok {
		return caller.call(from, protocol, a)
	}
	return nil, nil
}
