package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/pkg/code"
	"github.com/bvisness/bvisness.me/pkg/config"
	"github.com/bvisness/bvisness.me/pkg/images"
)

var hash string = fmt.Sprintf("%d", time.Now().Unix())

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to read build info")
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			hash = setting.Value
		}
	}

	time.Local = time.UTC
}

var bvisnessIncludes = &bhp.FSSearcher{
	FS: os.DirFS("include"),
}

func main() {
	b := bhp.Instance{
		SrcDir:     "site",
		FourOhFour: "404.luax",
		Searchers: []bhp.Searcher{
			bhp.GoSearcher{
				"images": images.LoadLib,
				"code":   code.LoadLib,
			},
			bvisnessIncludes,
		},
		StaticPaths: []string{"apps/"},
		Middleware:  bhp.ChainMiddleware(images.Middleware),

		Dev: config.Config.Dev,
	}
	b.Run()
}
