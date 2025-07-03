/**
 * @file Luax grammar for tree-sitter
 * @author Ben Visness <ben@bvisness.me>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

/*
 * Modeled after the official Lua 5.2 grammar and parser:
 *
 * https://www.lua.org/manual/5.2/manual.html#9
 * https://www.lua.org/source/5.2/lparser.c.html
 * 
 * Note that the parser actually disagrees pretty substantially with the
 * grammar, so that's cool. The parser source is annotated with grammar rules
 * to make it about as easy to follow as the EBNF grammar.
 */

module.exports = grammar({
  name: "luax",

  rules: {
    source_file: $ => optional($._block),

    _block: $ => repeat1($._stat),

    _stat: $ => choice(
      ";",
      $.ifstat,
      $.whilestat,
      $.dostat,
      $.forstat,
      $.repeatstat,
      $.funcstat,
      $.localstat,
      $.label,
      $.retstat,
      $.breakstat,
      $.gotostat,
      $.exprstat,
    ),

    ifstat: $ => seq(
      "if", $._subexpr, "then",
      optional($._block),
      repeat(seq(
        "elseif", $._subexpr, "then",
        optional($._block),
      )),
      optional(seq(
        "else",
        optional($._block),
      )),
      "end",
    ),

    whilestat: $ => seq(
      "while", $._subexpr, "do",
      optional($._block),
      "end",
    ),

    dostat: $ => seq(
      "do",
      optional($._block),
      "end",
    ),

    forstat: $ => seq(
      "for",
      choice(
        seq( // fornum
          $.name, "=", $._subexpr, ",", $._subexpr, optional(seq(",", $._subexpr)),
        ),
        seq( // forlist
          $.name, repeat(seq(",", $.name)), "in", $._explist,
        ),
      ),
      "do",
      optional($._block),
      "end",
    ),

    repeatstat: $ => seq(
      "repeat",
      optional($._block),
      "until",
      $._subexpr,
    ),

    funcstat: $ => seq("function", $.funcname, $._body),
    funcname: $ => seq(
      $.name, repeat($._fieldsel), prec(1, optional(seq(":", alias($.name, $.method_name)))),
    ),
    _body: $ => seq(
      "(", $.parlist, ")",
      optional($._block),
      "end",
    ),

    localstat: $ => seq(
      "local",
      choice(
        seq("function", $.name, $._body),
        seq(
          $.name, repeat(seq(",", $.name)),
          optional(seq("=", $._explist)),
        ),
      ),
    ),

    label: $ => seq("::", $.name),
    retstat: $ => seq("return", $._explist),
    breakstat: $ => "break",
    gotostat: $ => seq("goto", $.name),
    
    exprstat: $ => seq(
      $._suffixedexp,
      repeat(seq(",", $._suffixedexp)),
      optional(seq("=", $._explist)),
    ),
    
    _explist: $ => seq(
      $._subexpr, repeat(seq(",", $._subexpr)),
    ),
    parlist: $ => seq(
      $._param, repeat(seq(",", $._param)),
    ),

    _fieldsel: $ => seq(choice(".", ":"), $.name),
    _param: $ => choice($.name, "..."),

    _primaryexp: $ => choice($.name, seq("(", $._subexpr, ")")),
    _suffixedexp: $ => prec.left(seq(
      $._primaryexp,
      repeat(choice(
        seq(".", $.name),
        seq("[", $._subexpr, "]"),
        seq(":", $.name, $.funcargs),
        $.funcargs,
      )),
    )),
    _subexpr: $ => prec.left(seq(
      choice(
        seq($.unop, $._subexpr),
        seq($._simpleexp)
      ),
      repeat(seq($.binop, $._subexpr)),
    )),
    _simpleexp: $ => choice(
      $.number, $.string, "nil", "true", "false", "...",
      $.constructor_,
      $.tag,
      seq("function", $._body),
      $._suffixedexp,
    ),

    funcargs: $ => choice(
      seq("(", optional($._explist), ")"),
      $.constructor_,
      $.string,
    ),

    constructor_: $ => seq(
      "{",
      optional(seq(
        $.field,
        repeat(seq(choice(",", ";"), $.field)),
        optional(choice(",", ";")),
      )),
      "}",
    ),
    field: $ => choice(
      $._subexpr,
      seq(
        choice($.name, seq("[", $._subexpr, "]")), "=", $._subexpr,
      ),
    ),

    name: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,
    string: $ => /("([^"\\]|\\.)*"|'([^'\\]|\\.)*')/,
    number: $ => /0[xX][0-9a-fA-F]+\.?[0-9a-fA-F]*[pP]?[+-]?\d*|\d+\.?\d*(?:[eE][+-]?\d*)?|\.\d+(?:[eE][+-]?\d*)?/,

    unop: $ => choice("-", "not", "#"),
    binop: $ => choice(
      "+", "-", "*", "/", "^", "%",
      "..",
      "<", "<=", ">", ">=", "==", "~=",
      "and", "or",
    ),

    tag: $ => choice(
      seq("<!DOCTYPE", $.name, ">"),
      seq(
        "<",
        choice(
          ">",
          "/>",
          seq(
            $.name, ">",
          ),
        ),
      ),
    ),
  },
});
