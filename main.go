package main

import (
	"time"

	"github.com/bvisness/bvisness.me/bhp"
)

type Bvisness struct {
	Articles []Article
}

type HeaderData struct {
	Title string
}

type Article struct {
	HeaderData
	Date    time.Time
	Slug    string
	Excerpt string
	Url     string
}

func main() {
	bhp.Run("site", "include", Bvisness{
		Articles: []Article{
			{
				HeaderData: HeaderData{
					Title: "Untangling a bizarre WASM crash in Chrome",
				},
				Date: time.Date(2021, 7, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	})
}
