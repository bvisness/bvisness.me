module github.com/bvisness/bvisness.me

go 1.19

require (
	github.com/alecthomas/chroma/v2 v2.7.0
	github.com/chai2010/webp v1.1.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/stretchr/testify v1.8.0
	github.com/yuin/gopher-lua v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/yuin/gopher-lua v1.1.0 => github.com/bvisness/gopher-lua v0.0.0-20231210210735-90501ab9848b
