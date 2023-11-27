function Video(atts)
    return {
        type = "html",
        name = "div",
        atts = {
            class = "relative aspect-ratio--16x9",
        },
        children = {
            { type = "source", file = "test/video.luax", 74,  84 },
            {
                type = "html",
                name = "video",
                atts = {
                    class = "aspect-ratio--object",
                    src = relurl("vids/" .. atts.slug .. ".mp4"),
                    poster = relurl("vids/" .. atts.slug .. ".jpg"),
                    autoplay = true,
                    muted = true,
                    loop = true,
                    controls = true,
                    preload = "metadata",
                },
                children = { len = 0 },
            },
            { type = "source", file = "test/video.luax", 350, 356 },
            len = 3
        },
    }
end

function Wide(atts, children)
    if #children ~= 2 then
        error("requires exactly two children")
    end

    return {
        type = "html",
        name = "div",
        atts = {
            class = "wide flex justify-center mv4",
        },
        children = {
            { type = "source", file = "test/video.luax", 542, 552 },
            {
                type = "html",
                name = "div",
                atts = {
                    class = {
                        "flex flex-column flex-row-l",
                        atts.class or "items-center",
                        "g4"
                    },
                },
                children = {
                    { type = "source", file = "test/video.luax", 683, 697 },
                    {
                        type = "html",
                        name = "div",
                        atts = { class = "w-100 flex-fair-l p-dumb", },
                        children = {
                            { type = "source", file = "test/video.luax", 735, 753 },
                            children[1],
                            { type = "source", file = "test/video.luax", 768, 782 },
                            len = 3
                        },
                    },
                    { type = "source", file = "test/video.luax", 788, 802 },
                    {
                        type = "html",
                        name = "div",
                        atts = { class = "w-100 flex-fair-l p-dumb", },
                        children = {
                            { type = "source", file = "test/video.luax", 840, 858 },
                            children[2],
                            { type = "source", file = "test/video.luax", 873, 887 },
                            len = 3
                        },
                    },
                    { type = "source", file = "test/video.luax", 893, 903 },
                    len = 5
                },
            },
            { type = "source", file = "test/video.luax", 909, 915 },
            len = 3
        },
    }
end

bhp.render({
    type = "fragment",
    children = {
        { type = "source", file = "test/video.luax", 943,  949 },
        Wide({}, {
            { type = "source", file = "test/video.luax", 955,  965 },
            {
                type = "fragment",
                children = {
                    { type = "source", file = "test/video.luax", 967,  981 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Before we go further, let me introduce you to programming in Dreams.",
                            { type = "source", file = "test/video.luax", 984, 1052 }, -- avoid allocating and escaping big strings by slicing from source
                            len = 1
                        },
                    },
                    { type = "source", file = "test/video.luax", 1056, 1072 },
                    Video({ slug = "wowow", }, { len = 0 }),
                    { type = "source", file = "test/video.luax", 1094, 1110 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Dreams code is made up of nodes and wires...",
                            { type = "source", file = "test/video.luax", 1113, 1473 },
                            len = 1
                        },
                    },
                    { type = "source", file = "test/video.luax", 1477, 1487 },
                    len = 7
                },
            },
            { type = "source", file = "test/video.luax", 1490, 1500 },
            Video({ slug = "basics", }, { len = 0 }),
            { type = "source", file = "test/video.luax", 1523, 1529 },
            len = 5
        }),
        { type = "source", file = "test/video.luax", 1536, 1538 },
        len = 3
    },
})
