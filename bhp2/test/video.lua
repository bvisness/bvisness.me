function Video(atts)
    return {
        type = "html",
        name = "div",
        atts = {
            class = "relative aspect-ratio--16x9",
        },
        children = {
            { type = "source", 74,  84 },
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
                children = {},
            },
            { type = "source", 350, 356 },
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
            { type = "source", 542, 552 },
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
                    { type = "source", 683, 697 },
                    {
                        type = "html",
                        name = "div",
                        atts = { class = "w-100 flex-fair-l p-dumb", },
                        children = {
                            { type = "source", 735, 753 },
                            children[1],
                            { type = "source", 768, 782 },
                        },
                    },
                    { type = "source", 788, 802 },
                    {
                        type = "html",
                        name = "div",
                        atts = { class = "w-100 flex-fair-l p-dumb", },
                        children = {
                            { type = "source", 840, 858 },
                            children[2],
                            { type = "source", 873, 887 },
                        },
                    },
                    { type = "source", 893, 903 },
                },
            },
            { type = "source", 909, 915 },
        },
    }
end

bhp.render({
    type = "fragment",
    children = {
        { type = "source", 943,  949 },
        Wide({}, {
            { type = "source", 955,  965 },
            {
                type = "fragment",
                children = {
                    { type = "source", 967,  981 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Before we go further, let me introduce you to programming in Dreams.",
                            { type = "source", 984, 1052 }, -- avoid allocating and escaping big strings by slicing from source
                        },
                    },
                    { type = "source", 1056, 1072 },
                    Video({ slug = "wowow", }, {}),
                    { type = "source", 1094, 1110 },
                    {
                        type = "html",
                        name = "p",
                        atts = {},
                        children = {
                            -- "Dreams code is made up of nodes and wires...",
                            { type = "source", 1113, 1473 },
                        },
                    },
                    { type = "source", 1477, 1487 },
                },
            },
            { type = "source", 1490, 1500 },
            Video({ slug = "basics", }, {}),
            { type = "source", 1523, 1529 },
        }),
        { type = "source", 1536, 1538 },
    },
})
