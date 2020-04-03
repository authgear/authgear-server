package webapp

type StateStore interface {
	Get(id string) (*State, error)
	Set(state *State) error
}
