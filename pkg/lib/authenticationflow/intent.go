package authenticationflow

import "reflect"

type Intent interface {
	Kinder
	InputReactor
}

func IntentKind(intent Intent) string {
	intentType := reflect.TypeOf(intent).Elem()
	return intentType.Name()
}
