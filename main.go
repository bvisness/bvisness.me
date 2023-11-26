package main

import (
	"fmt"
	"html/template"
	"os"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp2"
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

var bvisnessIncludes = bhp2.FSSearcher{
	FS: os.DirFS("include"),
}

func main() {
	b := bhp2.Instance{
		SrcDir:      "site",
		FSSearchers: []bhp2.FSSearcher{bvisnessIncludes},
		StaticPaths: []string{"apps/"},
		// Middleware:  bhp2.ChainMiddleware(images.Middleware[Bvisness]),
	}
	b.Run()

	// bhp2.Options[Bvisness]{
	// TODO: Dunno if this is necessary any more.
	// StaticPaths: []string{"apps/"},
	// Funcs: func(b bhp.Instance[Bvisness], r bhp.Request[Bvisness]) template.FuncMap {
	// 	return bhp.MergeFuncMaps(
	// 		images.TemplateFuncs(b, r),
	// 		markdown.TemplateFuncs,
	// 		template.FuncMap{
	// 			"article": func(slug string) Article {
	// 				for _, article := range articles {
	// 					if article.Slug == slug {
	// 						return article
	// 					}
	// 				}
	// 				panic(fmt.Errorf("No article found with slug %s", slug))
	// 			},
	// 			"bust": func(resourceUrl string) string {
	// 				resUrlParsed, err := url.Parse(resourceUrl)
	// 				if err != nil {
	// 					panic(err)
	// 				}
	// 				q := resUrlParsed.Query()
	// 				q.Set("v", hash)
	// 				resUrlParsed.RawQuery = q.Encode()
	// 				return resUrlParsed.String()
	// 			},
	// 			"permalink": func() string {
	// 				return bhp.RelURL(r.R, "/")
	// 			},

	// 			// Desmos article
	// 			"threegraph": func(js string) template.HTML {
	// 				result := template.HTML(bhp.Eval(r.T, "desmos/threegraph.html", Threegraph{
	// 					ID: r.User.Desmos.NextThreegraphID,
	// 					JS: template.JS(js),
	// 				}))
	// 				r.User.Desmos.NextThreegraphID++
	// 				return result
	// 			},
	// 			"desmos": func(opts template.JS, js string) template.HTML {
	// 				result := template.HTML(bhp.Eval(r.T, "desmos/desmos.html", Desmos{
	// 					ID:   r.User.Desmos.NextDesmosID,
	// 					Opts: opts,
	// 					JS:   template.JS(js),
	// 				}))
	// 				r.User.Desmos.NextDesmosID++
	// 				return result
	// 			},
	// 		},
	// 	)
	// },
	// 	Middleware: bhp2.ChainMiddleware(images.Middleware[Bvisness]),
	// },
}
