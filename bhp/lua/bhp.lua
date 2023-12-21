require("pprint")

bhp = {
    _sources = {},
    _instance = nil, -- userdata, *bhp.Instance
    _request = nil,  -- userdata, *http.Request
}

-- Void elements, i.e. those which can have no children and therefore have no
-- closing tag. To be used as a lookup table.
local void = {
    area = true,
    base = true,
    br = true,
    col = true,
    embed = true,
    hr = true,
    img = true,
    input = true,
    link = true,
    meta = true,
    param = true,
    source = true,
    track = true,
    wbr = true,
}

---@param b StringBuilder
local function renderRec(node, b)
    if node == nil then
        return
    end

    if type(node) ~= "table" then
        b:add(tostring(node))
        return
    end

    if node.type == "custom" then
        renderRec(node.func(node.atts, node.children), b)
    elseif node.type == "html" then
        b:add("<")
        b:add(node.name)
        for att, value in pairs(node.atts) do
            b:add(" ")

            if type(value) == "string" then
                local escaped = value:gsub('"', "&quot;")

                b:add(att)
                b:add("=\"")
                b:add(escaped)
                b:add("\"")
            elseif type(value) == "boolean" then
                if att then
                    b:add(att)
                end
            elseif type(value) == "number" then
                b:add(att)
                b:add("=\"")
                b:add(tostring(value))
                b:add("\"")
            elseif type(value) == "table" then
                if att ~= "class" then
                    error("only `class` can use a table for its value")
                end

                b:add(att)
                b:add("=\"")

                -- numeric fields; unconditionally added first
                for i, class in ipairs(value) do
                    if i > 1 then
                        b:add(" ")
                    end
                    b:add(class)
                end

                -- TODO: string fields, added if value is true

                b:add("\"")
            else
                error(string.format("unknown type %s for tag attribute %s", type(value), att))
            end
        end
        b:add(">")

        for i = 1, node.children.len or #node.children do
            renderRec(node.children[i], b)
        end

        local hasChildren = node.children.len or #node.children
        if hasChildren or not void[node.name] then
            b:add("</")
            b:add(node.name)
            b:add(">")
        end
    elseif node.type == "fragment" then
        for i = 1, node.children.len or #node.children do
            renderRec(node.children[i], b)
        end
    elseif node.type == "source" then
        local raw = bhp._sources[node.file]
        if not raw then
            pprint(bhp._sources)
            error("could not find source file '" .. node.file .. "'")
        end
        b:add(raw:sub(node[1] + 1, node[2]))
    elseif node.type == "doctype" then
        b:add("<!DOCTYPE html>")
    elseif node.type == nil then
        pprint(node)
        b:add("[ERROR: nil node type, see console]")
    else
        error(string.format("unknown luax node type '%s'", node.type))
    end
end

function bhp.render(node)
    local b = StringBuilder:new()
    renderRec(node, b)

    return b:toString()
end

function bhp.redirect(url, code)
    return {
        action = "redirect",
        url = url,
        code = code or 301,
    }
end

function bhp.response(opts, content)
    return {
        action = "full-response",
        code = opts.code or 200,
        headers = opts.headers or {},
        content = content,
    }
end

---@param nodes table[]
function bhp.nosource(nodes)
    local res = {}
    for i, node in ipairs(nodes) do
        if node.type ~= "source" then
            table.insert(res, node)
        end
    end
    return res
end

function bhp.expand(nodes)
    return {
        type = "fragment",
        children = nodes,
    }
end

function bhp.map(t, f)
    local res = {}
    for i, v in ipairs(t) do
        res[i] = f(v)
    end
    return bhp.expand(res)
end

function bhp.join(t, sep)
    if t.type == "fragment" then
        t = t.children
    end

    local res = {}
    local len = t.len or #t
    for i = 1, len - 1 do
        table.insert(res, t[i])
        table.insert(res, sep)
    end
    if len > 0 then
        table.insert(res, t[len])
    end
    return bhp.expand(res)
end

---
--- StringBuilder
---

---@class StringBuilder
---@field strs string[]
StringBuilder = {}

---@return StringBuilder
function StringBuilder:new()
    local instance = {
        strs = {},
    }
    setmetatable(instance, self)
    self.__index = self
    return instance
end

---@param str string
function StringBuilder:add(str)
    table.insert(self.strs, str)
end

---@return string
function StringBuilder:toString()
    local size = #self.strs
    local chunkSize = 1000 -- Adjust the chunk size based on your scenario
    local result = {}

    for i = 1, size, chunkSize do
        local chunkEnd = math.min(i + chunkSize - 1, size)
        table.insert(result, table.concat(self.strs, "", i, chunkEnd))
    end

    return table.concat(result)
end

return bhp
