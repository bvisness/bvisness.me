package main

import (
	"fmt"
	"html/template"
	"os"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/pkg/code"
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

type Bvisness struct {
	BaseData // shut up, errors

	Desmos DesmosData
}

type BaseData struct {
	Title          string
	Description    string
	OpenGraphImage string // Relative URL within site folder
	Banner         string // Relative URL within site folder
	BannerScale    int    // e.g. 2 for a 2x resolution source image
	LightOnly      bool
}

type Article struct {
	BaseData
	Date time.Time
	Slug string
	Url  string
}

type DesmosData struct {
	NextThreegraphID int
	NextDesmosID     int
}

type Threegraph struct {
	ID int
	JS template.JS
}

type Desmos struct {
	ID   int
	Opts template.JS
	JS   template.JS
}

var bvisnessIncludes = bhp.FSSearcher{
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
	}
	b.Run()
}
