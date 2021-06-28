package cobraviper

import (
	"fmt"

	"github.com/spf13/cobra"
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

func NewBinder() *Binder {
	return &Binder{
		Viper: viper.New(),
	}
}

type Binder struct {
	Viper *viper.Viper
}

func (b *Binder) BindString(flagSet *pflag.FlagSet, arg *StringArgument) {
	_ = flagSet.VarPF(NewStringValue(arg.DefaultValue), arg.ArgumentName, arg.Short, arg.Usage)
	if arg.EnvName != "" {
		_ = b.Viper.BindEnv(arg.ArgumentName, arg.EnvName)
	}
}

func (b *Binder) GetString(cmd *cobra.Command, arg *StringArgument) string {
	val, _ := b.GetRequiredString(cmd, arg)
	return val
}

func (b *Binder) GetRequiredString(cmd *cobra.Command, arg *StringArgument) (string, error) {
	flag := cmd.Flags().Lookup(arg.ArgumentName)
	if flag == nil {
		return "", fmt.Errorf("flag not found")
	}
	val := flag.Value.String()
	if val != "" {
		return val, nil
	}
	val = b.Viper.GetString(arg.ArgumentName)
	if val != "" {
		return val, nil
	}
	return "", fmt.Errorf("%s is required", arg.ArgumentName)
}

type StringArgument struct {
	ArgumentName string
	EnvName      string
	Short        string
	Usage        string
	DefaultValue string
}
