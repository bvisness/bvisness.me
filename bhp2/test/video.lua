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
                            { type = "source", file = "test/video.luax", 770, 784 },
                            len = 3
                        },
                    },
                    { type = "source", file = "test/video.luax", 790, 804 },
                    {
                        type = "html",
                        name = "div",
                        atts = { class = "w-100 flex-fair-l p-dumb", },
                        children = {
                            { type = "source", file = "test/video.luax", 842, 860 },
                            children[2],
                            { type = "source", file = "test/video.luax", 877, 891 },
                            len = 3
                        },
                    },
                    { type = "source", file = "test/video.luax", 897, 907 },
                    len = 5
                },
            },
            { type = "source", file = "test/video.luax", 913, 919 },
            len = 3
        },
    }
end

bhp.render({
    type = "fragment",
    children = {
        { type = "source", file = "test/video.luax", 947,  953 },
        Wide({}, {
            { type = "source", file = "test/video.luax", 959,  969 },
            {
                type = "fragment",
                children = {
                    { type = "source", file = "test/video.luax", 971,  985 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Before we go further, let me introduce you to programming in Dreams.",
                            { type = "source", file = "test/video.luax", 988, 1056 }, -- avoid allocating and escaping big strings by slicing from source
                            len = 1
                        },
                    },
                    { type = "source", file = "test/video.luax", 1060, 1076 },
                    Video({ slug = "wowow", }, { len = 0 }),
                    { type = "source", file = "test/video.luax", 1098, 1114 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Dreams code is made up of nodes and wires...",
                            { type = "source", file = "test/video.luax", 1117, 1477 },
                            len = 1
                        },
                    },
                    { type = "source", file = "test/video.luax", 1481, 1491 },
                    len = 7
                },
            },
            { type = "source", file = "test/video.luax", 1494, 1504 },
            Video({ slug = "basics", }, { len = 0 }),
            { type = "source", file = "test/video.luax", 1527, 1533 },
            len = 5
        }),
        { type = "source", file = "test/video.luax", 1540, 1542 },
        len = 3
    },
})
