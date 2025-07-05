(ifstat ["if" "then" "elseif" "else" "end"] @keyword)
(whilestat ["while" "do" "end"] @keyword)
(dostat ["do" "end"] @keyword)
(forstat ["for" "in" "do" "end"] @keyword)
(repeatstat ["repeat" "until"] @keyword)
(funcstat
  ["function" "end"] @keyword
  (funcname) @function
)
(localstat
  ["local" "function"] @keyword
  (name) @vardef
)
(retstat ["return"] @keyword)
(breakstat) @keyword
(gotostat ["goto"] @keyword)
(exprstat) ; TODO

(htmlcomment) @comment
[(binop) (unop)] @operator

(params (name) @parameter)
(suffixedexp (name) @variable)
(getfield (name) @property)
