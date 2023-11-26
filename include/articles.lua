local function slugify(articles)
    local res = {}
    for _, article in ipairs(articles) do
        if article.slug == nil then
            error("article missing slug!")
        end
        res[article.slug] = article
    end
    return res
end

return slugify({
    {
        title = "“You can’t do that because I hate you.”",
        description = "An infuriating pattern that devs need to stop.",
        slug = "you-cant",
        date = os.time({ year = 2023, month = 7, day = 26 }),
    },
    {
        title = "Coroutines make robot code easy",
        description =
        "Our FIRST Robotics team struggled with autonomous code for years. Coroutines were the missing piece.",
        slug = "coroutines",
        date = os.time({ year = 2023, month = 6, day = 19 }),
    },
    {
        title = "How (not) to write a manifesto",
        description =
        "The Handmade Manifesto is on its third revision now. Let's look back at old versions of the manifesto and see how our messaging has shifted over time.",
        slug = "manifesto",
        date = os.time({ year = 2023, month = 5, day = 19 }),
    },
    {
        title = "\"It's always a tradeoff\"",
        description =
        "Programmers love to say things like \"it all depends\" or \"it's always a tradeoff\". This makes them sound very wise, but it's usually a cop-out.",
        slug = "tradeoffs",
        date = os.time({ year = 2023, month = 4, day = 15 }),
    },
    {
        title = "I did Advent of Code on a PlayStation 5 (take that hacker news)",
        description = "How far can I get in Advent of Code if I do all the problems in Dreams?",
        opengraphImage = "advent-of-dreams/vids/crane.jpg",
        slug = "advent-of-dreams",
        date = os.time({ year = 2022, month = 12, day = 31 }),
    },
    {
        title = "Essential complexity does not exist",
        description =
        "Trying to define \"essential complexity\" is a waste of time, but maybe not for the reason you think.",
        banner = "essential-complexity/gears.png",
        bannerScale = 3,
        slug = "essential-complexity",
        date = os.time({ year = 2022, month = 10, day = 15 }),
    },
    {
        title = "Untangling a bizarre WASM crash in Chrome",
        description = "How we solved a strange issue involving the guts of Chrome and the Go compiler.",
        opengraphImage = "chrome-wasm-crash/ogimage.png",
        slug = "chrome-wasm-crash",
        date = os.time({ year = 2021, month = 7, day = 9 }),
    },
    {
        title = "How to make a 3D renderer in Desmos",
        description =
        "Learn about the math of 3D rendering, and how to convince a 2D graphing calculator to produce 3D images.",
        opengraphImage = "desmos/opengraph.png",
        lightOnly = true,
        slug = "desmos",
        date = os.time({ year = 2019, month = 4, day = 14 }),
    },
    {
        title = "UE4: How to Make Awesome Buttons in VR",
        description = "Or: why the physics engine is not your friend.",
        banner = "vr-buttons/mediamenu.jpg",
        bannerScale = 2,
        slug = "vr-buttons",
        date = os.time({ year = 2017, month = 8, day = 27 }),
    },
    {
        title = "Blender masking layers: a quick tutorial",
        description = "A long response to a short StackExchange question.",
        slug = "blender-masking-layers",
        date = os.time({ year = 2017, month = 4, day = 25 }),
    },
    {
        title = "UE4: Controlling Spotify in-game",
        description =
        "And iTunes, Windows Media Player, and everything else, with just a little bit of Windows API magic.",
        banner = "ue4-spotify/mediamenu.jpg",
        bannerScale = 2,
        slug = "ue4-spotify",
        date = os.time({ year = 2017, month = 2, day = 12 }),
    },
    {
        title = "Compiling and using libgit2",
        description =
        "How to build libgit2 from source, install it on your computer, and use it in a project without linker errors.",
        slug = "libgit2",
        date = os.time({ year = 2017, month = 1, day = 2 }),
    },
    {
        title = "Project spotlight: VRInteractions",
        description = "An engine plugin for Unreal Engine 4 that makes it easy to create interactive objects in VR.",
        slug = "vrinteractions",
        date = os.time({ year = 2016, month = 11, day = 7 }),
    },
})
