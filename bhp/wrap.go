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
