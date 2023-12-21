local articles = require("articles")

return {
  type = "fragment",
  children = {
    [[<?xml version="1.0" standalone="yes" ?>]],
    { type = "source", file = "test/xml.luax", 96,  100 },
    {
      type = "html",
      name = "rss",
      atts = {
        version = "2.0",
        ["xmlns:atom"] = "http://www.w3.org/2005/Atom",
      },
      children = {
        { type = "source", file = "test/xml.luax", 160, 166 },
        {
          type = "html",
          name = "channel",
          atts = {},
          children = {
            { type = "source", file = "test/xml.luax", 175, 183 },
            {
              type = "html",
              name = "title",
              atts = {},
              children = {
                { type = "source", file = "test/xml.luax", 190, 201 },
                len = 1
              },
            },
            { type = "source", file = "test/xml.luax", 209, 217 },
            {
              type = "html",
              name = "link",
              atts = {},
              children = {
                { type = "source", file = "test/xml.luax", 223, 243 },
                len = 1
              },
            },
            { type = "source", file = "test/xml.luax", 250, 258 },
            {
              type = "html",
              name = "description",
              atts = {},
              children = {
                { type = "source", file = "test/xml.luax", 271, 298 },
                len = 1
              },
            },
            { type = "source", file = "test/xml.luax", 312, 320 },
            {
              type = "html",
              name = "language",
              atts = {},
              children = {
                { type = "source", file = "test/xml.luax", 330, 335 },
                len = 1
              },
            },
            { type = "source", file = "test/xml.luax", 346, 356 },
            {
              type = "html",
              name = "atom:link",
              atts = {
                href = "https://bvisness.me/index.xml",
                rel = "self",
                type = "application/rss+xml",
              },
              children = { len = 0 },
            },
            { type = "source", file = "test/xml.luax", 444, 454 },
            bhp.map(articles, function(a)
              return {
                type = "html",
                name = "item",
                atts = {},
                children = {
                  { type = "source", file = "test/xml.luax", 510, 522 },
                  {
                    type = "html",
                    name = "title",
                    atts = {},
                    children = {
                      a.title,
                      len = 1
                    },
                  },
                  { type = "source", file = "test/xml.luax", 550, 562 },
                  {
                    type = "html",
                    name = "link",
                    atts = {},
                    children = {
                      absurl(a.slug),
                      len = 1
                    },
                  },
                  { type = "source", file = "test/xml.luax", 595, 607 },
                  {
                    type = "html",
                    name = "pubDate",
                    atts = {},
                    children = {
                      os.date("%a, %d %b %Y %H:%M:%S %z", a.date),
                      len = 1
                    },
                  },
                  { type = "source", file = "test/xml.luax", 675, 689 },
                  {
                    type = "html",
                    name = "guid",
                    atts = {},
                    children = {
                      absurl(a.slug),
                      len = 1
                    },
                  },
                  { type = "source", file = "test/xml.luax", 722, 734 },
                  {
                    type = "html",
                    name = "description",
                    atts = {},
                    children = {
                      a.description,
                      len = 1
                    },
                  },
                  { type = "source", file = "test/xml.luax", 780, 790 },
                  len = 11
                },
              }
            end),
            { type = "source", file = "test/xml.luax", 812, 820 },
            len = 13
          },
        },
        { type = "source", file = "test/xml.luax", 830, 834 },
        len = 3
      },
    },
    { type = "source", file = "test/xml.luax", 840, 842 },
    len = 4
  },
}
