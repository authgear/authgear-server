package cobraviper

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
