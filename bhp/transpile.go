package bhp

import (
	"fmt"
	"strconv"
	"strings"
)

func SafeName(filename string) string {
	return strings.ReplaceAll(filename, "\\", "/")
}

func Transpile(source, filename string) (string, error) {
	tr := Transpiler{source: source, filename: filename}
	tr.skipWhitespace()
	tr.parseBlock()
	tr.expect(eof)
	tr.b.WriteString(tr.source[tr.luaChunkStart:])
	return tr.b.String(), nil
}

// See the full Lua grammar:
// https://www.lua.org/manual/5.2/manual.html#9

// TODO: Track newlines, transpile with respect to newlines so that line numbers are preserved
type Transpiler struct {
	source, filename string
	cur, lastCur     int

	inHTML        bool
	luaChunkStart int

	whitespaceStart int
	indent          string

	b strings.Builder
}

func (t *Transpiler) switchToHTML() {
	if !t.inHTML {
		t.b.WriteString(t.source[t.luaChunkStart:t.cur])
		t.luaChunkStart = -1
	}
	t.inHTML = true
}

func (t *Transpiler) switchToLua() {
	if t.inHTML {
		t.luaChunkStart = t.cur
	}
	t.inHTML = false
}

func (t *Transpiler) fail(msg string, args ...any) error {
	newArgs := []any{t.filename, t.lastCur}
	newArgs = append(newArgs, args...)
	return fmt.Errorf("bad LuaX syntax in %s at %d: "+msg, newArgs...)
}

var operators = []string{
	// longest first!

	"!", "--", // for the transpiler specifically

	"...",
	"..",
	".",

	"+", "-", "*", "/", "%", "^", "#",
	"==", "~=", "<=", ">=", "<", ">", "=",
	"(", ")", "{", "}", "[", "]", "::",
	";", ":", ",",
}

var eof = ""

func isalpha(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isdigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isspace(c byte) bool {
	return c == ' ' || c == '\f' || c == '\n' || c == '\r' || c == '\t' || c == '\v'
}

func isName(tok string) bool {
	if tok == eof {
		return false
	}
	return isalpha(tok[0]) || tok[0] == '_'
}

func isNumber(tok string) bool {
	if tok == eof {
		return false
	}
	return isdigit(tok[0])
}

func isString(tok string) bool {
	if tok == eof {
		return false
	}
	return tok[0] == '\'' || tok[0] == '"' || tok[0] == '['
}

func isBinop(tok string) bool {
	switch tok {
	case "+", "-", "*", "/", "^", "%", "..", "<", "<=", ">", ">=", "==", "~=", "and", "or":
		return true
	default:
		return false
	}
}

func isUnop(tok string) bool {
	switch tok {
	case "-", "not", "#":
		return true
	default:
		return false
	}
}

func (t *Transpiler) nextIs(s string) bool {
	return len(t.source[t.cur:]) >= len(s) && t.source[t.cur:t.cur+len(s)] == s
}

func (t *Transpiler) skipWhitespace() {
	t.whitespaceStart = t.cur
	indentStart := -1
	for {
		if t.cur >= len(t.source) {
			break
		} else if !t.inHTML && t.nextIs("--") {
			// comment! possibly multiline
			t.cur += 2
			long := t.nextIs("[[")
			if long {
				t.cur += 2
			}

			for {
				if !long && t.source[t.cur] == '\n' {
					t.cur += 1
					break
				}
				if long && t.nextIs("]]") {
					t.cur += 2
					break
				}
				t.cur++
			}
		} else if isspace(t.source[t.cur]) {
			t.cur++
			if t.source[t.cur-1] == '\n' {
				indentStart = t.cur
			}
		} else {
			break
		}
	}
	if indentStart >= 0 {
		t.indent = t.source[indentStart:t.cur]
	}
}

func (t *Transpiler) unwindWhitespace() {
	t.cur = t.whitespaceStart
}

func (t *Transpiler) lexName() string {
	start, end := t.cur, t.cur
	for {
		if end >= len(t.source) {
			break
		}
		c := t.source[end]
		isIdentChar := isalpha(c) || isdigit(c) || c == '_'
		if !isIdentChar {
			break
		}
		end++
	}
	return t.source[start:end]
}

func (t *Transpiler) lexNumber() string {
	start, end := t.cur, t.cur
	if t.nextIs("0x") || t.nextIs("0X") {
		// hex mode
		end += 2
		for {
			if end >= len(t.source) {
				break
			}
			c := t.source[end]
			if c == '.' || isdigit(c) || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F') {
				end++
			} else if c == 'p' || c == 'P' {
				end++
				if t.source[end] == '-' || t.source[end] == '+' {
					end++
				}
			} else {
				break
			}
		}
	} else {
		for {
			if end >= len(t.source) {
				break
			}
			c := t.source[end]
			if c == '.' || isdigit(c) {
				end++
			} else if c == 'e' || c == 'E' {
				end++
				if t.source[end] == '-' || t.source[end] == '+' {
					end++
				}
			} else {
				break
			}
		}
	}
	return t.source[start:end]
}

func (t *Transpiler) lexString(delim byte) string {
	if t.cur+1 >= len(t.source) {
		// weird edge case of source ending in quote
		// TODO: Error on EOF
		return ""
	}

	start, end := t.cur, t.cur+1
	escaping := false
	for {
		if end >= len(t.source) {
			// TODO: Error on EOF
			break
		}
		if escaping {
			end++
			escaping = false
			continue
		}

		if t.source[end] == '\\' {
			escaping = true
			end++
		} else if t.source[end] == delim {
			end++
			break
		} else {
			end++
		}
	}

	result := t.source[start:end]
	return result
}

func (t *Transpiler) lexLongString() string {
	start, end := t.cur, t.cur+1
	for {
		if end+1 >= len(t.source) {
			// TODO: Error on EOF
			break
		}
		if t.source[end] == ']' && t.source[end+1] == ']' {
			end += 2
			break
		}
		end++
	}
	result := t.source[start:end]
	return result
}

func (t *Transpiler) peekToken() string {
	// very bad hack
	defer func(cur int) {
		t.cur = cur
	}(t.cur)
	t.skipWhitespace()

	if t.cur >= len(t.source) {
		return eof
	}

	c := t.source[t.cur]
	if isalpha(c) || c == '_' {
		return t.lexName()
	}
	if isdigit(c) {
		return t.lexNumber()
	}
	if c == '\'' || c == '"' {
		return t.lexString(c)
	}
	if t.nextIs("[=") {
		panic(t.fail("shut up nerd"))
	}
	if t.nextIs("[[") {
		return t.lexLongString()
	}
	if len(t.source[t.cur:]) >= 2 && t.source[t.cur] == '.' && isdigit(t.source[t.cur+1]) {
		return "." + t.lexNumber()
	}
	for _, op := range operators {
		if t.nextIs(op) {
			return op
		}
	}
	panic(t.fail("wat is token I do not know (char: %c)", c))
}

func (t *Transpiler) peekToken2() string {
	// godless hack
	originalCur := t.cur
	t.nextToken()
	res := t.peekToken()
	t.cur = originalCur
	return res
}

func (t *Transpiler) peekTokenN(n int) string {
	// hack enshrined as feature
	originalCur := t.cur
	for i := 0; i < n-1; i++ {
		t.nextToken()
	}
	res := t.peekToken()
	t.cur = originalCur
	return res
}

func (t *Transpiler) nextToken() string {
	t.skipWhitespace()
	tok := t.peekToken()
	t.lastCur = t.cur
	t.cur += len(tok)
	t.skipWhitespace()
	return tok
}

func (t *Transpiler) expect(s string) {
	tok := t.nextToken()
	if tok != s {
		panic(t.fail("expected %s but got %s", s, tok))
	}
}

func (t *Transpiler) expectName(desc string) string {
	tok := t.nextToken()
	if !isName(tok) {
		panic(t.fail("expected name %s but got %s", desc, tok))
	}
	return tok
}

func (t *Transpiler) expectTagOrAttName(desc string) string {
	name := t.expectName(desc)
	for t.peekToken() == "-" || t.peekToken() == ":" {
		name += t.nextToken()
		name += t.expectName(desc)
	}
	return name
}

func (t *Transpiler) maybe(s string) {
	if t.peekToken() == s {
		t.nextToken()
	}
}

func (t *Transpiler) parseStat() (isLast bool) {
	switch tok := t.peekToken(); tok {
	case ";":
		t.nextToken()
	case "if":
		t.parseCondAndBlock()
		for t.peekToken() == "elseif" {
			t.parseCondAndBlock()
		}
		if t.peekToken() == "else" {
			t.nextToken()
			t.parseBlock()
		}
		t.expect("end")
		return false
	case "while":
		t.nextToken()
		t.parseSubexp()
		t.expect("do")
		t.parseBlock()
		t.expect("end")
		return false
	case "do":
		panic(t.fail("do statements are not implemented"))
	case "for":
		t.nextToken()
		t.expectName("of loop variable")
		switch tok := t.peekToken(); tok {
		case "=":
			t.expect("=")
			t.parseSubexp()
			t.expect(",")
			t.parseSubexp()
			if t.peekToken() == "," {
				t.parseSubexp()
			}
		case ",", "in":
			for t.peekToken() == "," {
				t.nextToken()
				t.expectName("of loop variable")
			}
			t.expect("in")
			t.parseExpList()
		default:
			panic(t.fail("unexpected token in loop: %s", tok))
		}
		t.expect("do")
		t.parseBlock()
		t.expect("end")
		return false
	case "repeat":
		panic(t.fail("not implemented"))
	case "function":
		t.nextToken()
		t.parseFuncName()
		t.parseFuncBody()
		return false
	case "local":
		t.nextToken()
		if t.peekToken() == "function" {
			t.nextToken()
			t.expectName("of local function")
			t.parseFuncBody()
		} else {
			// LOCAL NAME {`,' NAME} [`=' explist]
			t.expectName("of local var")
			for t.peekToken() == "," {
				t.nextToken()
				t.expectName("of local var")
			}
			if t.peekToken() == "=" {
				t.nextToken()
				t.parseExpList()
			}
		}
		return false
	case "break":
		t.nextToken() // skip BREAK
		return true
	case "::", "return", "goto":
		panic(t.fail(tok + " is not implemented"))
	default:
		t.parseExprStat()
		return false
	}

	panic("unreachable")
}

func (t *Transpiler) parseCondAndBlock() {
	t.nextToken()   // if | elseif
	t.parseSubexp() // condition
	t.expect("then")
	// TODO: something weird about goto? see test_then_block in lparser.c
	t.parseBlock()
}

func (t *Transpiler) parseExprStat() {
	t.parseSuffixedExp()
	if t.peekToken() == "=" || t.peekToken() == "," {
		t.parseAssignment()
	}
}

func (t *Transpiler) parseAssignment() {
	if t.peekToken() == "," {
		t.nextToken()
		t.parseSuffixedExp()
		t.parseAssignment()
	} else {
		t.expect("=")
		t.parseExpList()
	}
}

func (t *Transpiler) parseFuncName() {
	// Name {'.' Name} [':' Name]
	t.expectName("of function")
	for {
		if t.peekToken() != "." {
			break
		}
		t.nextToken()
		t.expectName("in function name")
	}
	if t.peekToken() == ":" {
		t.nextToken()
		t.expectName("in function name")
	}
}

// Includes the parenthesized list of arguments
func (t *Transpiler) parseFuncBody() {
	// '(' [parlist] ')'
	t.expect("(")
	for {
		if t.peekToken() == ")" {
			t.nextToken()
			break
		}

		name := t.nextToken()
		if !isName(name) && name != "..." {
			panic(t.fail("expected argument name for function but got %s", name))
		}
		if t.peekToken() == "," {
			t.nextToken()
			continue
		}

		t.expect(")")
		break
	}

	t.parseBlock()
	t.expect("end")
}

func (t *Transpiler) parseBlock() {
	for {
		switch tok := t.peekToken(); tok {
		case "else", "elseif", "end", eof, "until":
			return
		}

		if t.peekToken() == "return" {
			t.nextToken()
			t.parseExpList()
			t.maybe(";")
			break
		} else {
			isLast := t.parseStat()
			t.maybe(";")
			if isLast {
				break
			}
		}
	}
}

func (t *Transpiler) parseExpList() {
	t.parseSubexp()
	for {
		if t.peekToken() != "," {
			break
		}
		t.nextToken()
		t.parseSubexp()
	}
}

func (t *Transpiler) parsePrimaryExp() {
	// NAME | '(' expr ')'
	switch tok := t.peekToken(); tok {
	case "(":
		t.nextToken()
		t.parseSubexp()
		t.nextToken()
	default:
		if isName(tok) {
			t.nextToken()
			return
		}
		panic(t.fail("unexpected token in primary exp: %s", tok))
	}
}

func (t *Transpiler) parseSuffixedExp() {
	// primaryexp { '.' NAME | '[' exp ']' | ':' NAME funcargs | funcargs }
	t.parsePrimaryExp()
	for {
		switch tok := t.peekToken(); tok {
		case ".":
			t.nextToken()
			t.expectName("of field")
		case "[":
			t.nextToken()
			t.parseSubexp()
			t.expect("]")
		case ":":
			t.nextToken()
			t.expectName("of method")
			t.parseFuncArgs()
		case "(", "{":
			t.parseFuncArgs()
		default:
			if isString(tok) {
				t.parseFuncArgs()
				return
			}
			return
		}
	}
}

func (t *Transpiler) parseFuncArgs() {
	switch tok := t.peekToken(); tok {
	case "(":
		t.nextToken()
		if t.peekToken() == ")" {
			t.nextToken()
			return
		}
		t.parseExpList()
		t.expect(")")
	case "{":
		t.parseTable()
	default:
		if isString(tok) {
			t.nextToken()
			return
		}
		panic(t.fail("unknown token in func args: %s", tok))
	}
}

func (t *Transpiler) parseSubexp() {
	tok := t.peekToken()
	if isUnop(tok) {
		t.nextToken()
		t.parseSubexp()
	} else {
		t.parseSimpleExp()
	}

	for isBinop(t.peekToken()) {
		t.nextToken()
		t.parseSubexp()
	}
}

func (t *Transpiler) parseSimpleExp() {
	switch tok := t.peekToken(); tok {
	case "nil", "true", "false", "...":
		t.nextToken()
	case "{":
		t.parseTable()
	case "<":
		t.parseTag(t.indent, true)
	case "function":
		t.nextToken()
		t.parseFuncBody()
	default:
		if isNumber(tok) || isString(tok) {
			t.nextToken()
			return
		}
		t.parseSuffixedExp()
	}
}

func (t *Transpiler) parseTable() {
	t.expect("{")

	for {
		if t.peekToken() == "}" {
			t.nextToken()
			break
		}

		if t.peekToken2() == "=" {
			// Name '=' exp
			t.expectName("for table field")
			t.nextToken() // '='
			t.parseSubexp()
		} else if t.peekToken() == "[" {
			// '[' exp ']' '=' exp
			t.nextToken()
			t.parseSubexp()
			t.expect("]")
			t.nextToken() // '='
			t.parseSubexp()
		} else {
			// exp
			t.parseSubexp()
		}

		if t.peekToken() == "," || t.peekToken() == ";" {
			t.nextToken()
			continue
		}

		t.expect("}")
		break
	}
}

type TagAttribute struct {
	Name  string
	Value string
}

// finally the actual transpiler part
func (t *Transpiler) parseTag(indent string, fromLua bool) {
	if fromLua {
		t.switchToHTML()
		defer t.switchToLua()
	}

	t.expect("<")
	if t.peekToken() == ">" {
		// fragment
		t.expect(">")
		t.unwindWhitespace()

		t.b.WriteString("{\n")
		t.b.WriteString("    ")
		t.b.WriteString(`type = "fragment",` + "\n")
		t.b.WriteString(indent)
		t.b.WriteString("    ")
		t.b.WriteString("children = ")

		t.parseTagChildren("", true, indent+"    ")

		t.b.WriteString(",\n")
		t.b.WriteString(indent)
		t.b.WriteString("}")
	} else if t.peekToken() == "!" {
		t.expect("!")
		specialName := t.expectName("of special tag")
		if specialName != "DOCTYPE" {
			panic(t.fail("only DOCTYPE is supported for special tags"))
		}

		doctype := t.expectName("for doctype")
		if doctype != "html" {
			panic(t.fail("`html` is the only supported doctype"))
		}
		t.expect(">")

		t.b.WriteString(`{ type = "doctype" }`)
	} else {
		// named tag
		tagName := t.expectTagOrAttName("of tag")
		var atts []TagAttribute
		for isName(t.peekToken()) {
			att := TagAttribute{Name: t.expectTagOrAttName("of attribute")}
			if t.peekToken() == "=" {
				t.nextToken()
				if isString(t.peekToken()) {
					att.Value = t.nextToken()
				} else if t.peekToken() == "{" {
					t.nextToken()
					expStart := t.cur
					t.parseSubexp()
					expEnd := t.cur
					t.expect("}")
					att.Value = strings.TrimSpace(t.source[expStart:expEnd])
				}
			} else {
				att.Value = "true"
			}
			atts = append(atts, att)
		}

		isCustom := 'A' <= tagName[0] && tagName[0] <= 'Z'

		t.b.WriteString("{\n")
		t.b.WriteString(indent)
		t.b.WriteString("    ")
		if isCustom {
			t.b.WriteString(`type = "custom",` + "\n")
			t.b.WriteString(indent)
			t.b.WriteString("    ")
			t.b.WriteString("func = " + tagName + ",\n")
		} else {
			t.b.WriteString(`type = "html",` + "\n")
			t.b.WriteString(indent)
			t.b.WriteString("    ")
			t.b.WriteString("name = \"" + tagName + "\",\n")
		}
		t.b.WriteString(indent)
		t.b.WriteString("    ")
		t.b.WriteString("atts = ")

		if len(atts) > 0 {
			t.b.WriteString("{\n")
			for _, att := range atts {
				t.b.WriteString(indent)
				t.b.WriteString("    ")
				t.b.WriteString("    ")
				if strings.Contains(att.Name, "-") || strings.Contains(att.Name, ":") {
					t.b.WriteString(`["`)
					t.b.WriteString(att.Name)
					t.b.WriteString(`"]`)
				} else {
					t.b.WriteString(att.Name)
				}
				t.b.WriteString(" = ")
				t.b.WriteString(att.Value)
				t.b.WriteString(",\n")
			}
			t.b.WriteString(indent)
			t.b.WriteString("    ")
			t.b.WriteString("},\n")
		} else {
			t.b.WriteString("{},\n")
		}

		hasChildren := true
		if t.peekToken() == "/" {
			t.nextToken()
			hasChildren = false
		}
		t.expect(">")
		t.unwindWhitespace()

		t.b.WriteString(indent)
		t.b.WriteString("    ")
		t.b.WriteString("children = ")
		if hasChildren {
			allowTags := !(tagName == "script" || tagName == "style")
			t.parseTagChildren(tagName, allowTags, indent+"    ")
		} else {
			t.b.WriteString("{ len = 0 }")
		}
		t.b.WriteString(",\n")
		t.b.WriteString(indent)
		t.b.WriteString("}")
	}

	t.luaChunkStart = t.cur
}

func (t *Transpiler) parseTagChildren(tagName string, allowTags bool, indent string) {
	t.b.WriteString("{\n")

	textStart := t.cur
	n := 0
	for {
		if t.nextIs("{{") {
			t.emitTextNode(textStart, indent+"    ", &n)

			t.b.WriteString(indent)
			t.b.WriteString("    ")

			t.expect("{")
			t.expect("{")
			t.switchToLua()
			t.parseSubexp()
			t.switchToHTML()
			t.expect("}")
			if !t.nextIs("}") { // dumb hack to avoid making }} a token
				t.expect("}}")
			}
			t.expect("}")
			t.unwindWhitespace()

			t.b.WriteString(",\n")

			textStart = t.cur
			n += 1
			continue
		}

		// Closing-tag behavior differs when we allow tags vs. not.
		// When tags are allowed, a `<` always means something taggy.
		// When they're not, we require more exact closing behavior.
		var isOpening, isClosing bool
		if allowTags {
			isOpening = t.nextIs("<")
			isClosing = t.nextIs("<") && t.peekToken2() == "/"
		} else {
			isClosing = t.nextIs("<") && t.peekToken2() == "/" && isName(t.peekTokenN(3)) && t.peekTokenN(3) == tagName && t.peekTokenN(4) == ">"
		}
		isComment := t.nextIs("<") && t.peekToken2() == "!" && t.peekTokenN(3) == "--"

		if isClosing {
			// closing tag
			t.emitTextNode(textStart, indent+"    ", &n)

			t.expect("<")
			t.expect("/")
			if tagName != "" {
				name := t.expectName("of closing tag")
				if name != tagName {
					panic(t.fail("expected </%s> but got </%s>", tagName, name))
				}
			}
			t.expect(">")
			t.unwindWhitespace()
			break
		} else if isComment {
			// comment
			t.emitTextNode(textStart, indent+"    ", &n)

			t.expect("<")
			t.expect("!")
			t.expect("--")
			for t.source[t.cur:t.cur+3] != "-->" {
				t.cur++
			}
			t.cur += 3

			textStart = t.cur
		} else if isOpening {
			// opening tag
			t.emitTextNode(textStart, indent+"    ", &n)

			t.b.WriteString(indent)
			t.b.WriteString("    ")
			t.parseTag(indent+"    ", false)
			t.b.WriteString(",\n")

			n += 1

			textStart = t.cur
		} else {
			t.cur++
		}
	}

	t.b.WriteString(indent)
	t.b.WriteString("len = ")
	t.b.WriteString(strconv.Itoa(n))

	t.b.WriteString(indent)
	t.b.WriteString("}")
}

func (t *Transpiler) emitTextNode(start int, indent string, numChildren *int) {
	if start == t.cur {
		return
	}
	t.b.WriteString(indent)
	t.b.WriteString(`{ type = "source", file = "`)
	t.b.WriteString(t.filename)
	t.b.WriteString(`", `)
	t.b.WriteString(strconv.Itoa(start))
	t.b.WriteString(", ")
	t.b.WriteString(strconv.Itoa(t.cur))
	t.b.WriteString(" },\n")
	*numChildren += 1
}
