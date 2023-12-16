require("common")
require("article/title")

function SimpleArticle(atts, children)
    if not atts.article then
        return {
            type = "custom",
            func = Error,
            atts = {},
            children = {
                { type = "source", file = "test/children.luax", 139, 169 },
                len = 1
            },
        }
    end

    return {
        type = "custom",
        func = Base,
        atts = {},
        children = {
            { type = "source", file = "test/children.luax", 207, 217 },
            {
                type = "custom",
                func = Common,
                atts = {},
                children = {
                    { type = "source", file = "test/children.luax", 225, 239 },
                    {
                        type = "html",
                        name = "article",
                        atts = {},
                        children = {
                            { type = "source", file = "test/children.luax", 248, 266 },
                            {
                                type = "html",
                                name = "header",
                                atts = {},
                                children = {
                                    { type = "source", file = "test/children.luax", 274, 296 },
                                    {
                                        type = "custom",
                                        func = ArticleTitle,
                                        atts = {
                                            article = atts.article,
                                        },
                                        children = { len = 0 },
                                    },
                                    { type = "source", file = "test/children.luax", 337, 355 },
                                    len = 3
                                },
                            },
                            { type = "source", file = "test/children.luax", 364, 382 },
                            children,
                            { type = "source", file = "test/children.luax", 396, 410 },
                            len = 5
                        },
                    },
                    { type = "source", file = "test/children.luax", 420, 430 },
                    len = 3
                },
            },
            { type = "source", file = "test/children.luax", 439, 445 },
            len = 3
        },
    }
end
