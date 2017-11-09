package bi

//Session Session
type Session interface {
	GetSessionImpl() *SessionImpl
}
