require("common")
require("article/title")

function SimpleArticle(atts, children)
    if not atts.article then
        return Error({}, {
            { type = "source", file = "test/children.luax", 139, 169 },
            len = 1
        })
    end

    return Base(
        {},
        {
            { type = "source", file = "test/children.luax", 207, 217 },
            Common(
                {},
                {
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
                                    ArticleTitle(
                                        {
                                            article = atts.article,
                                        },
                                        { len = 0 }
                                    ),
                                    { type = "source", file = "test/children.luax", 337, 355 },
                                    len = 3
                                },
                            },
                            { type = "source", file = "test/children.luax", 364, 382 },
                            children,
                            { type = "source", file = "test/children.luax", 394, 408 },
                            len = 5
                        },
                    },
                    { type = "source", file = "test/children.luax", 418, 428 },
                    len = 3
                }
            ),
            { type = "source", file = "test/children.luax", 437, 443 },
            len = 3
        }
    )
end
