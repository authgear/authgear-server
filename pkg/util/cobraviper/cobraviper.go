package cobraviper

import (
	"fmt"
	"strconv"

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

func (b *Binder) BindInt(flagSet *pflag.FlagSet, arg *IntArgument) {
	_ = flagSet.VarPF(NewIntValue(arg.DefaultValue), arg.ArgumentName, arg.Short, arg.Usage)
	if arg.EnvName != "" {
		_ = b.Viper.BindEnv(arg.ArgumentName, arg.EnvName)
	}
}

func (b *Binder) GetString(cmd *cobra.Command, arg *StringArgument) string {
	val, _ := b.GetRequiredString(cmd, arg)
	return val
}

func (b *Binder) GetInt(cmd *cobra.Command, arg *IntArgument) (*int, error) {
	validate := func(val int) error {
		if arg.Max != nil && val > *arg.Max {
			return fmt.Errorf(
				"%s should be smaller than or equal to %d",
				arg.ArgumentName,
				*arg.Max,
			)
		}

		if arg.Min != nil && val < *arg.Min {
			return fmt.Errorf(
				"%s should be larger than or equal to %d",
				arg.ArgumentName,
				*arg.Min)
		}
		return nil
	}

	flag := cmd.Flags().Lookup(arg.ArgumentName)
	if flag == nil {
		return nil, fmt.Errorf("flag not found")
	}

	intVal, ok := flag.Value.(*IntValue)
	if ok && intVal.Val != nil {
		if err := validate(*intVal.Val); err != nil {
			return nil, err
		}
		return intVal.Val, nil
	}

	strVal := b.Viper.GetString(arg.ArgumentName)
	if strVal != "" {
		intVal, err := strconv.Atoi(strVal)
		if err != nil {
			return nil, fmt.Errorf("%s is not an integer", arg.ArgumentName)
		}
		if err := validate(intVal); err != nil {
			return nil, err
		}
		return &intVal, nil
	}

	return nil, nil
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

func (b *Binder) GetRequiredInt(cmd *cobra.Command, arg *IntArgument) (int, error) {
	val, err := b.GetInt(cmd, arg)
	if err != nil {
		return 0, err
	}
	if val == nil {
		return 0, fmt.Errorf("%s is required", arg.ArgumentName)
	}
	return *val, nil
}
