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

func tokens(source string) []string {
	tr := Transpiler{source: source}
	tr.skipWhitespace()
	var tokens []string
	for {
		tok := tr.nextToken()
		tokens = append(tokens, tok)
		if tok == eof {
			break
		}
	}
	return tokens
}

func TestTokenizer(t *testing.T) {
	for _, test := range tokenizerTests {
		t.Run(test.name, func(t *testing.T) {
			toks := tokens(test.source)
			assert.Equal(t, test.numTokens, len(toks))
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
	{"fancy example with root statements", `
function Wide(atts, children)
    if #children ~= 2 then
        error("requires exactly two children")
    end

    return __html("div", { class = "wide flex justify-center mv4" }, {
        __html("div", {
            class = {
                "flex flex-column flex-row-l",
                atts.class or "items-center",
                "g4"
            },
        }, {
            __html("div", { class = "w-100 flex-fair-l p-dumb" }, {
                children[1],
            }),
            __html("div", { class = "w-100 flex-fair-l p-dumb" }, {
                children[2],
            }),
        }),
    })
end

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
}))`},
	{"path stuff", `
---mirrors a path, does not make a copy
function Path:mirror()
	for _, point in ipairs(self.points) do
		point.x = 651.25 - point.x
	end

	self.startAngle = math.pi - self.startAngle
	self.endAngle = math.pi - self.endAngle
end

function Path:print()
	for _, point in ipairs(self.points) do
		print(point)
	end
end

test("Path:new", function(t)
	local p = Path:new("TestOnlyDoNotEdit", {
		testEvent = function()
			print("wow, an event!")
		end
	})
	t:assertEqual(#p.distances, #p.points, "we should have one distance for each point")
	t:assert(p.events[1].func ~= nil)
end)
	`},
	{"robot coroutines", `
if isTesting() then
	autoChooser = MockSendableChooser
end

local doNothingAuto = FancyCoroutine:new(function()
end)

local function sleep(timeSeconds)
	local timer = Timer:new()
	timer:start()

	while timer:get() < timeSeconds do
		coroutine.yield()
	end
end
	`},
}

func TestVanillaLua(t *testing.T) {
	for _, test := range vanillaParserTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Transpile(test.source)
			assert.Nil(t, err)
		})
	}
}

type TagTest struct {
	name     string
	source   string
	expected string
}

var tagTests = []TagTest{
	{
		"simple self-closing",
		`local tag = <div foo="bar" baz bing />`,
		`local tag = __tag("div", { foo="bar", baz=true, bing=true, }, {})`,
	},
	{
		"inline Lua expressions",
		`local tag = <div foo="bar" baz={ 1 + 2 } bing={ foo.bar:greet("hello") } />`,
		`local tag = __tag("div", { foo="bar", baz=1 + 2, bing=foo.bar:greet("hello"), }, {})`,
	},
	{
		"before/after",
		`
local tag1 = <foo />
local tag2 = <bar />
print("good :)")
`,
		`
local tag1 = __tag("foo", {}, {})
local tag2 = __tag("bar", {}, {})
print("good :)")
		`,
	},
}

func TestTags(t *testing.T) {
	for _, test := range tagTests {
		t.Run(test.name, func(t *testing.T) {
			transpiled, err := Transpile(test.source)
			if assert.Nil(t, err) {
				t.Log(transpiled)
				actualToks := tokens(transpiled)
				expectedToks := tokens(test.expected)
				assert.Equal(t, expectedToks, actualToks)
			}
		})
	}
}
