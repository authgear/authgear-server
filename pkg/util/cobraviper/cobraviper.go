package cobraviper

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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

func (arg *StringArgument) Bind(flagSet *pflag.FlagSet, v *viper.Viper) {
	_ = v.BindPFlag(arg.ArgumentName, flagSet.VarPF(NewStringValue(arg.DefaultValue), arg.ArgumentName, arg.Short, arg.Usage))
	_ = v.BindEnv(arg.ArgumentName, arg.EnvName)
}

func (arg *StringArgument) Get(v *viper.Viper) string {
	return v.GetString(arg.ArgumentName)
}

func (arg *StringArgument) GetRequired(v *viper.Viper) (string, error) {
	val := v.GetString(arg.ArgumentName)
	if val == "" {
		return "", fmt.Errorf("%s is required", arg.ArgumentName)
	}
	return val, nil
}
