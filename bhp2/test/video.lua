function Video(atts)
    return __tag("div", {
        class = "relative aspect-ratio--16x9",
    }, {
        __text(74, 84),
        __tag("video", {
            class = "aspect-ratio--object",
            src = relurl("vids/" .. atts.slug .. ".mp4"),
            poster = relurl("vids/" .. atts.slug .. ".jpg"),
            autoplay = true,
            muted = true,
            loop = true,
            controls = true,
            preload = "metadata",
        }, {}),
        __text(350, 356),
    })
end

function Wide(atts, children)
    if #children ~= 2 then
        error("requires exactly two children")
    end

    return __tag("div", {
        class = "wide flex justify-center mv4",
    }, {
        __text(542, 552),
        __tag("div", {
            class = {
                "flex flex-column flex-row-l",
                atts.class or "items-center",
                "g4"
            },
        }, {
            __text(683, 697),
            __tag("div", { class = "w-100 flex-fair-l p-dumb", }, {
                __text(735, 753),
                children[1],
                __text(768, 782),
            }),
            __text(788, 802),
            __tag("div", { class = "w-100 flex-fair-l p-dumb", }, {
                __text(840, 858),
                children[2],
                __text(873, 887),
            }),
            __text(893, 903),
        }),
        __text(909, 915),
    })
end

render(__fragment({
    __text(939, 945),
    __tag("Wide", {}, {
        __text(951, 961),
        __fragment({
            __text(963, 977),
            __tag("p", {}, {
                -- "Before we go further, let me introduce you to programming in Dreams.",
                __text(980, 1048), -- avoid allocating and escaping big strings by slicing from source
            }),
            __text(1052, 1068),
            __tag("Video", { slug = "wowow", }, {}),
            __text(1090, 1106),
            __tag("p", {}, {
                -- "Dreams code is made up of nodes and wires...",
                __text(1109, 1469),
            }),
            __text(1473, 1483),
        }),
        __text(1486, 1496),
        __tag("Video", { slug = "basics", }, {}),
        __text(1519, 1525),
    }),
    __text(1532, 1534),
}))
