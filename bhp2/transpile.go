package bhp2

import (
	"fmt"
	"regexp"
)

func Transpile(source string) (string, error) {
	tr := Transpiler{source: source}
	tr.skipWhitespace()
	tr.parseBlock()
	tr.expect(eof)
	return "so transpiley wow", nil
}

// See the full Lua grammar:
// https://www.lua.org/manual/5.2/manual.html#9

type Transpiler struct {
	source string
	cur    int
}

var operators = []string{
	// longest first!
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

var reWhitespace = regexp.MustCompile(`(\s|--(\[\[]]))*`)

func (t *Transpiler) nextIs(s string) bool {
	return len(t.source[t.cur:]) >= len(s) && t.source[t.cur:t.cur+len(s)] == s
}

func (t *Transpiler) skipWhitespace() {
	for {
		if t.cur >= len(t.source) {
			break
		} else if t.nextIs("--") {
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
		} else {
			break
		}
	}
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
		return ""
	}

	start, end := t.cur, t.cur+1
	escaping := false
	for {
		if end > len(t.source) {
			break
		}
		if escaping {
			end++
			continue
		}
		escaping = false

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
	start, end := t.cur, t.cur
	for {
		if end > len(t.source) {
			break
		}
		if t.nextIs("]]") {
			end += 2
			break
		}
		end++
	}
	result := t.source[start:end]
	return result
}

func (t *Transpiler) peekToken() string {
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
		panic("shut up nerd")
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
	panic("wat is token I do not know")
}

func (t *Transpiler) peekToken2() string {
	// godless hack
	originalCur := t.cur
	t.nextToken()
	res := t.peekToken()
	t.cur = originalCur
	return res
}

func (t *Transpiler) nextToken() string {
	tok := t.peekToken()
	t.cur += len(tok)
	t.skipWhitespace()
	return tok
}

func (t *Transpiler) expect(s string) {
	tok := t.nextToken()
	if tok != s {
		panic(fmt.Errorf("bad Lua syntax: expected %s but got %s", s, tok))
	}
}

func (t *Transpiler) expectName(desc string) {
	tok := t.nextToken()
	if !isName(tok) {
		panic(fmt.Errorf("expected name %s but got %s", desc, tok))
	}
}

func (t *Transpiler) maybe(s string) {
	if t.peekToken() == s {
		t.nextToken()
	}
}

func (t *Transpiler) parseStat() {
	switch tok := t.peekToken(); tok {
	case ";":
		t.nextToken()
	case "if":
		t.parseCondAndBlock()
		for t.peekToken() == "elseif" {
			t.parseCondAndBlock()
		}
		if t.peekToken() == "else" {
			t.parseBlock()
		}
		t.expect("end")
	case "while":
		t.nextToken()
		t.parseSubexp()
		t.expect("do")
		t.parseBlock()
		t.expect("end")
	case "do":
		panic("unimplemented")
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
			panic(fmt.Errorf("unexpected token in loop: %s", tok))
		}
		t.expect("do")
		t.parseBlock()
		t.expect("end")
	case "repeat":
		panic("not implemented")
	case "function":
		t.nextToken()
		t.parseFuncName()
		t.parseFuncBody()
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
	case "::", "return", "break", "goto":
		panic("unimplemented")
	default:
		t.parseExprStat()
	}
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
			panic(fmt.Errorf("expected argument name for function but got %s", name))
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
			t.parseStat()
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
		panic(fmt.Errorf("unexpected token in primary exp: %s", tok))
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
		panic(fmt.Errorf("unknown token in func args: %s", tok))
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
