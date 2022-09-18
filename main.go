package main

import (
	"fmt"
	"html/template"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/pkg/images"
	"github.com/bvisness/bvisness.me/pkg/markdown"
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
}

type Bvisness struct {
	Articles []Article
}

type BaseData struct {
	Title          string
	Description    string
	OpenGraphImage string // Relative URL within site folder
}

type CommonData struct {
	Banner string
}

type Article struct {
	BaseData
	CommonData
	Date time.Time
	Slug string
	Url  string
}

type UserData struct {
	Desmos DesmosData
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

var articles = []Article{
	{
		BaseData: BaseData{
			Title:          "Untangling a bizarre WASM crash in Chrome",
			Description:    "How we solved a strange issue involving the guts of Chrome and the Go compiler.",
			OpenGraphImage: "chrome-wasm-crash/ogimage.png",
		},
		Slug: "chrome-wasm-crash",
		Date: time.Date(2021, 7, 9, 0, 0, 0, 0, time.UTC),
	},
	{
		BaseData: BaseData{
			Title:          "How to make a 3D renderer in Desmos",
			Description:    "Learn about the math of 3D rendering, and how to convince a 2D graphing calculator to produce 3D images.",
			OpenGraphImage: "desmos/opengraph.png",
		},
		Slug: "desmos",
		Date: time.Date(2019, 4, 14, 0, 0, 0, 0, time.UTC),
	},
}

func main() {
	bhp.Run(
		"site", "include",
		UserData{
			Desmos: DesmosData{
				NextThreegraphID: 1,
			},
		},
		func(r bhp.Request[UserData]) template.FuncMap {
			return bhp.MergeFuncMaps(
				images.TemplateFuncs,
				markdown.TemplateFuncs,
				template.FuncMap{
					"article": func(slug string) Article {
						for _, article := range articles {
							if article.Slug == slug {
								return article
							}
						}
						panic(fmt.Errorf("No article found with slug %s", slug))
					},
					"bust": func(resourceUrl string) string {
						resUrlParsed, err := url.Parse(resourceUrl)
						if err != nil {
							panic(err)
						}
						q := resUrlParsed.Query()
						q.Set("v", hash)
						resUrlParsed.RawQuery = q.Encode()
						return resUrlParsed.String()
					},
					"permalink": func() string {
						return bhp.RelURL(r.R, "/")
					},

					// Desmos article
					"threegraph": func(js string) template.HTML {
						result := template.HTML(bhp.Eval(r.T, "desmos/threegraph.html", Threegraph{
							ID: r.User.Desmos.NextThreegraphID,
							JS: template.JS(js),
						}))
						r.User.Desmos.NextThreegraphID++
						return result
					},
					"desmos": func(opts template.JS, js string) template.HTML {
						result := template.HTML(bhp.Eval(r.T, "desmos/desmos.html", Desmos{
							ID:   r.User.Desmos.NextDesmosID,
							Opts: opts,
							JS:   template.JS(js),
						}))
						r.User.Desmos.NextDesmosID++
						return result
					},
				},
			)
		},
		Bvisness{
			Articles: articles,
		},
	)
}
