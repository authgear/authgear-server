//go:build zxcvbn_test
// +build zxcvbn_test

package password_test

import (
	"io/ioutil"
	"math/rand"
	"runtime"
	"testing"

	"github.com/lithdew/quickjs"
	"github.com/trustelem/zxcvbn"
)

func jsVM(fn func(*quickjs.Context)) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	vm := quickjs.NewRuntime()
	defer vm.Free()
	context := vm.NewContext()
	defer context.Free()

	js, err := ioutil.ReadFile("../../../../../resources/authgear/static/zxcvbn.js")
	if err != nil {
		panic(err)
	}
	value, err := context.Eval(string(js))
	if err != nil {
		panic(err)
	}
	value.Free()
	fn(context)
}

func zxcvbnBasic(input string) int {
	return zxcvbn.PasswordStrength(input, nil).Score
}

func TestZXCVBNCorrectnessRandom(t *testing.T) {
	jsVM(func(context *quickjs.Context) {
		r := rand.New(rand.NewSource(1234))
		randString := func() string {
			length := r.Intn(10) + 1
			str := make([]byte, length)
			for i := 0; i < length; i++ {
				str[i] = byte(r.Intn(96) + 32)
			}
			return string(str)
		}

		count := 1000
		for count > 0 {
			input := randString()
			check(context, input, nil, t)
			count--
		}

	})
}

func TestZXCVBNCorrectnessFixture(t *testing.T) {
	jsVM(func(context *quickjs.Context) {
		fixtures := []struct {
			input     string
			userInput []string
		}{
			{"abcde123456", nil},
			{"nihongo-wo-manabimashou", nil},
		}

		for _, c := range fixtures {
			check(context, c.input, c.userInput, t)
		}
	})
}

func check(context *quickjs.Context, input string, userInput []string, t *testing.T) {
	scoreActual := zxcvbnBasic(input)

	jsArray := context.Array()
	for i, s := range userInput {
		jsArray.SetByInt64(int64(i), context.String(s))
	}
	context.Globals().Set("userInput", jsArray)
	context.Globals().Set("input", context.String(input))
	value, err := context.Eval("zxcvbn(input, userInput)")
	if err != nil {
		panic(err)
	}
	defer value.Free()
	scoreExpected := value.Get("score").Int32()

	if scoreActual != int(scoreExpected) {
		t.Logf("actual zxcvbn: %s", string(value.String()))
		t.Fatalf("incorrect score for input string %q: (%d != %d)", input, scoreActual, scoreExpected)
	} else {
		t.Logf("%q = %d", input, scoreActual)
	}
}
