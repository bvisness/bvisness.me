(functioncall name: (prefixexp) @function)
(label) @label
(breakstat) @keyword
(gotostat
  "goto" @keyword
  (name) @label
)
(dostat ["do" "end"] @keyword)
(whilestat ["while" "do" "end"] @keyword)
(repeatstat ["repeat" "until"] @keyword)
(ifstat ["if" "then" "elseif" "else" "end"] @keyword)
(forstat ["for" "do" "end"] @keyword)
(forinstat ["for" "in" "do" "end"] @keyword)
(forinstat
  arg: (name) @variable.declaration
)
(funcstat ["function" "end"] @keyword)
(funcstat
  name: (funcname) @function
  par: (name) @parameter
)
(localfuncstat ["local" "function" "end"] @keyword)
(localfuncstat
  name: (name) @function
  par: (name) @parameter
)
(assignstat_local
  ["local"] @keyword
  lhs: (name) @variable.declaration
)
(retstat ["return"] @keyword)

(functiondef ["function" "end"] @keyword)
(functiondef
  par: (name) @parameter
)

[(binop) (unop)] @operator
(string) @string
(number) @number
"nil" @macro ; it looks good ok

(varname) @variable
(prefixexp (name) @variable)
(funcname [(name) (method_name)] @function)

(getprop (name) @property)
(tableconstructor ["{" "}"] @operator) ; ensure that we don't get bonked by {{ }} rules
(field (name) @property)

(functioncall name: (_) @function)

(htmlcomment) @comment
(att (name) @variable)
