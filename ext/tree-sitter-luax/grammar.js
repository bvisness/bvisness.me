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
    source_file: $ => optional($.block),

    block: $ => repeat1($._stat),

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
      "if", $.expr, "then",
      optional($.block),
      repeat(seq(
        "elseif", $.expr, "then",
        optional($.block),
      )),
      optional(seq(
        "else",
        optional($.block),
      )),
      "end",
    ),

    whilestat: $ => seq(
      "while", $.expr, "do",
      optional($.block),
      "end",
    ),

    dostat: $ => seq(
      "do",
      optional($.block),
      "end",
    ),

    forstat: $ => seq(
      "for",
      choice(
        seq( // fornum
          field("forarg", $.name),
          "=", field("formin", $.expr), ",", field("formax", $.expr), optional(seq(",", field("forstep", $.expr))),
        ),
        seq( // forlist
          field("forarg", $.name), repeat(seq(",", field("forarg", $.name))), "in", _explist($, "forexpr"),
        ),
      ),
      "do",
      optional($.block),
      "end",
    ),

    repeatstat: $ => seq(
      "repeat",
      optional($.block),
      "until",
      $.expr,
    ),

    funcstat: $ => seq("function", field("name", $.funcname), _body($)),
    funcname: $ => seq(
      $.name, repeat($._fieldsel), prec(1, optional(seq(":", alias($.name, $.method_name)))),
    ),
    params: $ => seq(
      "(",
      optional(seq(
        $._param, repeat(seq(",", $._param)),
      )),
      ")",
    ),

    localstat: $ => seq(
      "local",
      choice(
        seq("function", field("name", $.name), _body($)),
        seq(
          field("name", $.name), repeat(seq(",", field("name", $.name))),
          optional(seq("=", _explist($, "val"))),
        ),
      ),
    ),

    label: $ => seq("::", $.name),
    retstat: $ => seq("return", _explist($)),
    breakstat: $ => "break",
    gotostat: $ => seq("goto", $.name),
    
    exprstat: $ => seq(
      field("lhs", $.suffixedexp), repeat(seq(",", field("lhs", $.suffixedexp))),
      optional(seq("=", _explist($, "rhs"))),
    ),

    _fieldsel: $ => seq(choice(".", ":"), $.name),
    _param: $ => choice($.name, alias("...", $.ellipsis)),

    expr: $ => prec.left(seq(
      choice(
        seq($.unop, $.expr),
        seq($._simpleexp)
      ),
      repeat(seq($.binop, $.expr)),
    )),
    _primaryexp: $ => choice($.name, seq("(", $.expr, ")")),
    suffixedexp: $ =>prec.left(seq(
      $._primaryexp,
      repeat(choice(
        prec(1, seq($.getfield, $.funcargs)),
        $.getfield,
        prec(1, seq($.getindex, $.funcargs)),
        $.getindex,
        prec(1, seq($.getmethod, $.funcargs)),
        $.funcargs
      )),
    )),
    _simpleexp: $ => choice(
      $.number,
      $.string, "nil", "true", "false", "...",
      $.constructor_,
      $.tag,
      seq("function", _body($)),
      $.suffixedexp,
    ),
    getfield: $ => seq(".", $.name),
    getindex: $ => seq("[", $.expr, "]"),
    getmethod: $ => seq(":", $.name),

    funcargs: $ => choice(
      seq("(", optional(_explist($)), ")"),
      $.constructor_,
      $.string,
    ),

    constructor_: $ => seq(
      "{",
      optional(seq(
        $.fielddef,
        repeat(seq(choice(",", ";"), $.fielddef)),
        optional(choice(",", ";")),
      )),
      "}",
    ),
    fielddef: $ => choice(
      $.expr,
      seq(
        choice($.name, seq("[", $.expr, "]")), "=", $.expr,
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
      $.fragment,
      $.specialtag,
      $.namedtag,
    ),
    fragment: $ => seq("<>", $.tagchildren),
    specialtag: $ => seq("<!DOCTYPE", $.name, ">"),
    htmlcomment: $ => /<!--.*?-->/,
    namedtag: $ => seq(
      "<",
      choice(
        prec(1, seq(
          field("name", choice("script", "style")),
          repeat($.att),
          choice(
            "/>",
            seq(">", field("children", $.tagchildren_notags)),
          )
        )),
        seq(
          field("name", $.name),
          repeat($.att),
          choice(
            "/>",
            seq(">", field("children", $.tagchildren)),
          )
        ),
      ),
    ),

    att: $ => seq(
      $.name,
      optional(seq("=", choice(
        $.string,
        seq("{", $.expr, "}"),
      ))),
    ),

    tagchildren: $ => seq(
      repeat(choice(
        prec(1, seq("{{", $.expr, "}}")),
        prec(1, $.htmlcomment),
        prec(1, $.tag),
        "{", /[^<{]+/,
      )),
      "<", "/", alias(optional($.name), "closingname"), ">",
    ),
    tagchildren_notags: $ => seq(
      repeat(choice(
        prec(1, seq("{{", $.expr, "}}")),
        /[^<]+/,
        seq("<", /[^\/]/),
      )),
      "<", "/", alias(optional($.name), "closingname"), ">",
    ),
  },
});

function _body($) {
  return seq(
    field("params", $.params),
    field("body", optional($.block)),
    "end",
  );
}

function _explist($, fieldname) {
  return seq(
    fieldname ? field(fieldname, $.expr) : $.expr,
    repeat(seq(",", fieldname ? field(fieldname, $.expr) : $.expr)),
  );
}