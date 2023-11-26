require("pprint")

bhp = {
    _rendered = "",
    _sources = {},
}

---@param b StringBuilder
local function renderRec(node, b)
    if node == nil then
        return
    end

    if type(node) == "string" then
        b:add(node)
        return
    end

    if node.type == "html" then
        b:add("<")
        b:add(node.name)
        for att, value in pairs(node.atts) do
            b:add(" ")

            if type(value) == "string" then
                b:add(att)
                b:add("=\"")
                b:add(value)
                b:add("\"")
            elseif type(value) == "boolean" then
                if att then
                    b:add(att)
                end
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
                -- TODO: more specific error message
                error("unknown type for tag attribute")
            end
        end
        b:add(">")

        for i, child in ipairs(node.children) do
            renderRec(child, b)
        end

        b:add("</")
        b:add(node.name)
        b:add(">")
    elseif node.type == "fragment" then
        for i, child in ipairs(node.children) do
            renderRec(child, b)
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
        if #node > 0 then
            for i, child in ipairs(node) do
                renderRec(child, b)
            end
        else
            b:add("[ERROR: nil node type, see console]")
        end
    else
        error(string.format("unknown luax node type '%s'", node.type))
    end
end

function bhp.render(node)
    local b = StringBuilder:new()
    renderRec(node, b)

    bhp._rendered = b:toString()
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
    return table.concat(self.strs, "")
end

return bhp
