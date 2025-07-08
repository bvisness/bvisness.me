/**
 * @file Luax grammar for tree-sitter
 * @author Ben Visness <ben@bvisness.me>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

/*
 * Modeled after the official Lua 5.2 grammar:
 *
 * https://www.lua.org/manual/5.2/manual.html#9
 * 
 * The following Lua tree-sitter grammar was also consulted for help with
 * precedence and other nonsense:
 * 
 * https://github.com/tjdevries/tree-sitter-lua/blob/4932594a24f04e4ccf046919bc354272841b0077/grammar.js
 * 
 */

const PREC = {
  DEFAULT: 1,
  PRIORITY: 2,

  FUNCTION: 1,
  STATEMENT: 10,
};

module.exports = grammar({
  name: "luax",

  extras: $ => [/[\n]/, /\s/, $._comment],

  rules: {
    source_file: $ => optional($.block),

    block: $ => repeat1($._stat),

    _stat: $ => prec.right(PREC.STATEMENT, choice(
      ";",
      $.assignstat_global,
      $.functioncall,
      $.label,
      $.breakstat,
      $.gotostat,
      $.dostat,
      $.whilestat,
      $.repeatstat,
      $.ifstat,
      $.forstat,
      $.forinstat,
      $.funcstat,
      $.localfuncstat,
      $.assignstat_local,
      $.retstat,
    )),

    assignstat_global: $ => prec.right(PREC.DEFAULT, seq(
      list(() => field("lhs", $._var)), "=", list(() => field("rhs", $._exp)),
    )),

    label: $ => seq("::", $.name, "::"),
    breakstat: $ => "break",
    gotostat: $ => seq("goto", $.name),

    dostat: $ => seq(
      "do",
      optional($.block),
      "end",
    ),

    whilestat: $ => seq(
      "while", $._exp, "do",
      optional($.block),
      "end",
    ),

    repeatstat: $ => seq(
      "repeat",
      optional($.block),
      "until",
      $._exp,
    ),

    ifstat: $ => seq(
      "if", $._exp, "then",
      optional($.block),
      repeat(seq(
        "elseif", $._exp, "then",
        optional($.block),
      )),
      optional(seq(
        "else",
        optional($.block),
      )),
      "end",
    ),

    forstat: $ => seq(
      "for",
      field("arg", $.name), "=", field("min", $._exp), ",", field("max", $._exp),
      optional(seq(",", field("step", $._exp))),
      "do",
      optional($.block),
      "end",
    ),

    forinstat: $ => seq(
      "for",
      list(() => field("arg", $.name)), "in", list(() => field("exp", $._exp)),
      "do",
      optional($.block),
      "end",
    ),

    funcstat: $ => seq(
      "function", field("name", $.funcname), funcbody($),
    ),
    
    localfuncstat: $ => seq(
      "local", "function", field("name", $.name), funcbody($),
    ),

    assignstat_local: $ => prec.right(PREC.DEFAULT, seq(
      "local", list(() => field("lhs", $.name)), "=", list(() => field("rhs", $._exp)),
    )),

    retstat: $ => prec(PREC.PRIORITY, seq("return", list(() => $._exp))),
    
    funcname: $ => seq(
      $.name, repeat(seq(".", $.name)), optional(seq(":", alias($.name, $.method_name))),
    ),
    _var: $ => prec(PREC.PRIORITY, choice(
      alias($.name, $.varname),
      $.getindex,
      $.getprop,
    )),
    getindex: $ => seq($.prefixexp, "[", $._exp, "]"),
    getprop: $ => seq($.prefixexp, ".", $.name),
    
    _exp: $ => prec.left(choice(
      "nil", "false", "true",
      $.number, $.string,
      "...",
      $.functiondef,
      $.prefixexp,
      $.tableconstructor,
      seq($._exp, $.binop, $._exp),
      seq($.unop, $._exp),
      // New in LuaX
      $._tag,
    )),
    prefixexp: $ => choice(
      $._var,
      $.functioncall,
      seq("(", $._exp, ")"),
    ),

    functioncall: $ => prec.right(PREC.FUNCTION, seq(
      choice(
        field("name", $.prefixexp),
        seq($.prefixexp, ":", field("name", $.name)),
      ),
      field("args", $.args),
    )),
    args: $ => choice(
      seq("(", optional(list(() => $._exp)), ")"),
      $.tableconstructor,
      $.string,
    ),

    functiondef: $ => seq("function", funcbody($)),

    tableconstructor: $ => seq(
      "{",
      optional(list(() => $.field, () => choice(",", ";"), true)),
      "}",
    ),
    field: $ => choice(
      seq("[", $._exp, "]", "=", $._exp),
      seq($.name, "=", $._exp),
      $._exp,
    ),

    binop: $ => choice(
      "+", "-", "*", "/", "^", "%", "..",
      "<", "<=", ">", ">=", "==", "~=",
      "and", "or",
    ),
    unop: $ => choice("-", "not", "#"),

    name: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,
    string: $ => /("([^"\\]|\\.)*"|'([^'\\]|\\.)*')/,
    number: $ => /0[xX][0-9a-fA-F]+\.?[0-9a-fA-F]*[pP]?[+-]?\d*|\d+\.?\d*(?:[eE][+-]?\d*)?|\.\d+(?:[eE][+-]?\d*)?/,

    _tag: $ => choice(
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
        seq("{", $._exp, "}"),
      ))),
    ),

    tagchildren: $ => seq(
      repeat(choice(
        prec(1, seq("{{", $._exp, "}}")),
        prec(1, $.htmlcomment),
        prec(1, $._tag),
        "{", /[^<{]+/,
      )),
      "<", "/", alias(optional($.name), "closingname"), ">",
    ),
    tagchildren_notags: $ => seq(
      repeat(choice(
        prec(1, seq("{{", $._exp, "}}")),
        /[^<]+/,
        seq("<", /[^\/]/),
      )),
      "<", "/", alias(optional($.name), "closingname"), ">",
    ),

    _comment: $ => choice(
      /--[^\r\n]*/,
    )
  },
});

function funcbody($) {
  return seq(
    "(", optional(list(() => field("par", choice($.name, alias("...", $.ellipsis))))), ")",
    field("body", optional($.block)),
    "end",
  );
}

/** 
 * @param {function(): RuleOrLiteral} rule
 * @param {function(): RuleOrLiteral} sep
 */
function list(rule, sep = () => ",", trailing = false) {
  const seqMembers = [rule(), repeat(seq(sep(), rule()))];
  if (trailing) {
    seqMembers.push(optional(sep()));
  }
  return seq(...seqMembers);
}
