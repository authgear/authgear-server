package cobraviper

type StringValue string

func NewStringValue(val string) *StringValue {
	p := val
	return (*StringValue)(&p)
}

func (s *StringValue) Set(val string) error {
	*s = StringValue(val)
	return nil
}
func (s *StringValue) Type() string {
	return "string"
}

func (s *StringValue) String() string { return string(*s) }

type StringArgument struct {
	ArgumentName string
	EnvName      string
	Short        string
	Usage        string
	DefaultValue string
}
