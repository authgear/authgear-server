package common

// ExecError is error resulted from application logic of plugin (e.g.
// an exception thrown within a lambda function)
type ExecError struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

func (err *ExecError) Error() string {
	return err.Name + "\n" + err.Description
}
