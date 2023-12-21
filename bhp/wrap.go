package bhp

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func Raise(l *lua.LState, err error) int {
	l.Error(lua.LString(err.Error()), 1)
	return 0
}

func RaiseMsg(l *lua.LState, err error, msg string, args ...any) int {
	l.RaiseError("%s: %v", fmt.Sprintf(msg, args...), err.Error())
	return 0
}

func WrapS_S(f func(string) string) lua.LGFunction {
	return func(l *lua.LState) int {
		arg1 := l.ToString(1)
		res := f(arg1)
		l.Push(lua.LString(res))
		return 1
	}
}

func WrapSI_S(f func(string, int) string) lua.LGFunction {
	return func(l *lua.LState) int {
		arg1 := l.ToString(1)
		arg2 := l.ToInt(1)
		res := f(arg1, arg2)
		l.Push(lua.LString(res))
		return 1
	}
}

func WrapSI_SE(f func(string, int) (string, error)) lua.LGFunction {
	return func(l *lua.LState) int {
		arg1 := l.ToString(1)
		arg2 := l.ToInt(1)
		res, err := f(arg1, arg2)
		if err != nil {
			return Raise(l, err)
		}
		l.Push(lua.LString(res))
		return 1
	}
}
