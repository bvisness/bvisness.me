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

func checkString(l *lua.State, index int, what string) string {
	if s, ok := l.ToString(index); ok {
		return s
	}
	lua.ArgumentError(l, index, fmt.Sprintf("%s (string) expected, got %s", what, lua.TypeNameOf(l, index)))
	panic("unreachable")
}

func checkType(l *lua.State, index int, t lua.Type, what string) {
	if l.TypeOf(index) != t {
		lua.ArgumentError(l, index, fmt.Sprintf("%s (%s) expected, got %s", what, t.String(), lua.TypeNameOf(l, index)))
		panic("unreachable")
	}
}

func renderRec(l *lua.State, b *strings.Builder, source string) {
	lua.CheckType(l, -1, lua.TypeTable)

	l.Field(-1, "type")
	t := checkString(l, -1, "node type")
	l.Pop(1)

	switch t {
	case "html":
		l.Field(-1, "name")
		name := checkString(l, -1, "tag name")
		l.Field(-2, "atts")
		checkType(l, -1, lua.TypeTable, "tag attributes")
		l.Field(-3, "children")
		checkType(l, -1, lua.TypeTable, "tag children")

		b.WriteString("<")
		b.WriteString(name)
		l.PushNil()
		for l.Next(-3) { // atts
			b.WriteString(" ")

			att := checkString(l, -2, "attribute name")
			switch l.TypeOf(-1) {
			case lua.TypeString:
				val := checkString(l, -1, "attribute value")
				b.WriteString(att)
				b.WriteString(`="`)
				b.WriteString(val) // TODO: escape
				b.WriteString(`"`)
			case lua.TypeBoolean:
				has := l.ToBoolean(-1)
				if has {
					b.WriteString(att)
				}
			}
			l.Pop(1)

			// TODO: special handling of `class`
		}
		b.WriteString(">")

		l.PushNil()
		for l.Next(-2) { // children
			checkType(l, -1, lua.TypeTable, "tag child")
			renderRec(l, b, source)
			l.Pop(1)
		}

		b.WriteString("</")
		b.WriteString(name)
		b.WriteString(">")

		l.Pop(3) // name, atts, children
	case "fragment":
		l.Field(1, "children")
		checkType(l, -1, lua.TypeTable, "fragment children")

		l.PushNil()
		for l.Next(-2) {
			checkType(l, -1, lua.TypeTable, "fragment child")
			renderRec(l, b, source)
			l.Pop(1)
		}

		l.Pop(1) // children
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
