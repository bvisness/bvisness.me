package bhp2

import (
	"fmt"
	"strings"

	"github.com/Shopify/go-lua"
)

func render(l *lua.State) int {
	l.Global("bhp")
	l.Field(-1, "_source")
	source := lua.CheckString(l, -1)
	l.Pop(2)

	var b strings.Builder
	renderRec(l, &b, source)

	l.Global("bhp")
	l.PushString(b.String())
	l.SetField(-2, "_rendered")

	return 0
}

func renderRec(l *lua.State, b *strings.Builder, source string) {
	lua.CheckType(l, -1, lua.TypeTable)
	defer l.Pop(1)

	l.Field(-1, "type")
	t := lua.CheckString(l, -1)
	l.Pop(1)

	switch t {
	case "html":
		l.Field(1, "name")
		name := lua.CheckString(l, -1)
		l.Field(1, "atts")
		lua.CheckType(l, -1, lua.TypeTable)
		l.Field(1, "children")
		lua.CheckType(l, -1, lua.TypeTable)

		b.WriteString("<")
		b.WriteString(name)
		l.PushNil()
		for l.Next(-3) { // atts
			att := lua.CheckString(l, -2)
			val := lua.CheckString(l, -1)
			b.WriteString(" ")
			b.WriteString(att)
			b.WriteString(`="`)
			b.WriteString(val) // TODO: escape
			b.WriteString(`"`)
			l.Pop(1)

			// TODO: special handling of `class`
		}
		b.WriteString(">")

		l.PushNil()
		for l.Next(-2) { // children
			lua.CheckType(l, -1, lua.TypeTable)
			renderRec(l, b, source)
		}

		b.WriteString("</")
		b.WriteString(name)
		b.WriteString(">")

		l.Pop(3)
	case "source":
		l.RawGetInt(-1, 1)
		start := lua.CheckInteger(l, -1)
		l.RawGetInt(-2, 2)
		end := lua.CheckInteger(l, -1)
		b.WriteString(source[start:end])
		l.Pop(2)
	default:
		panic(fmt.Errorf("unknown luax node type '%s'", t))
	}
}

func getRendered(l *lua.State) string {
	l.Global("bhp")
	l.Field(-1, "_rendered")
	result := lua.CheckString(l, -1)
	l.Pop(2)
	return result
}

func setSource(l *lua.State, source string) {
	l.Global("bhp")
	l.PushString(source)
	l.SetField(-2, "_source")
	l.Pop(1)
}

var bhp2Library = []lua.RegistryFunction{
	{"render", render},
}

func BHP2Open(l *lua.State) int {
	lua.NewLibrary(l, bhp2Library)
	l.PushString("")
	l.SetField(-2, "_source")
	l.PushString("")
	l.SetField(-2, "_rendered")
	return 1
}
