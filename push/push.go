package push

// Mapper defines a single method Map()
type Mapper interface {
	// Implementor of Map should return a string-interface map which
	// all values are JSON-marshallable
	Map() map[string]interface{}
}

// MapMapper is a string-interface map that implemented the Mapper
// interface.
type MapMapper map[string]interface{}

// Map returns the map itself.
func (m MapMapper) Map() map[string]interface{} {
	return map[string]interface{}(m)
}

// Sender defines the methods that a push service should support.
type Sender interface {
	Send(m Mapper, deviceToken string) error
}
