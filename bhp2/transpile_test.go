package bhp2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tokenizerTest struct {
	name      string
	numTokens int
	source    string
}

var tokenizerTests = []tokenizerTest{
	{"numbers", 10, `
		3     3.0     3.1416     314.16e-2     0.31416E1
		0xff  0x0.1E  0xA23p-4   0X1.921FB54442D18P+1
	`},
	{"tables and functions", 76, `
		function Video(slug)
			return {
				__html("div", { class = "relative aspect-ratio--16x9" }, {
					__html("video", {
						class = "aspect-ratio--object",
						src = relurl("vids/" .. slug .. ".mp4"),
						poster = relurl("vids/" .. slug .. ".jpg"),
						autoplay = true,
						muted = true,
						loop = true,
						controls = true,
						preload = "metadata",
					})
				})
			}
		end
	`},
	{"now with comments", 57, `
		table.insert(__doc, Wide({}, {
			__fragment({
				__html("p", {
					-- "Before we go further, let me introduce you to programming in Dreams.",
					__source(123, 234), -- avoid allocating and escaping big strings by slicing from source
				}),
				Video("wowow"),
				__html("p", {
					-- "Dreams code is made up of nodes and wires...",
					__source(345, 456),
				}),
			}),
		}))
	`},
	{"beeg comment", 5, `
--[[

===============================================================================
								2023 LYON SPEC
===============================================================================

The Lyon system needs to always respect physical limits so that it does not run
into the ground or into the robot. These constraints must be maintained no
matter what we tell it to do in robot.lua.

The following functions must all be implemented and tested on the simulator
before running on the actual robot. For each, check the corresponding boxes
only when you have performed the specified tests.

-------------------------------------------------------------------------------
OVERVIEW

All the core logic of the core Lyon system will be performed in Lyon:periodic.
...

--]]

Lyon = {}
	`},
}

func TestTokenizer(t *testing.T) {
	for _, test := range tokenizerTests {
		t.Run(test.name, func(t *testing.T) {
			tr := Transpiler{source: test.source}
			tr.skipWhitespace()
			numTokens := 0
			for {
				tok := tr.nextToken()
				t.Log("torken", tok)
				numTokens++
				if tok == eof {
					break
				}
			}
			assert.Equal(t, test.numTokens, numTokens)
		})
	}
}

type parserTest struct {
	name   string
	source string
}

var vanillaParserTests = []parserTest{
	{"simple function w/ expression", `
		function foo.bar:baz(a, b)
			return a + b + (a - b)
		end
	`},
	{"assignment of table", `
		myTable = {
			a,
			1 + 2 - 3,
			foo = "bar",
			["baz"] = "I have been a good bing :)",
			[8] = 0xf00
		}
	`},
}

func TestTranspile(t *testing.T) {
	for _, test := range vanillaParserTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Transpile(test.source)
			assert.Nil(t, err)
		})
	}
}
