package tree_sitter_luax_test

import (
	"testing"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_luax "github.com/bvisness/bvisness.me/bindings/go"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_luax.Language())
	if language == nil {
		t.Errorf("Error loading LuaX grammar")
	}
}
