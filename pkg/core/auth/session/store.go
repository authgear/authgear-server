package session

type Store interface {
	Create(s *Session) error
	Update(s *Session) error
	Get(id string) (*Session, error)
	Delete(s *Session) error
}
