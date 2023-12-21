---@class Vec2
---@field x number
---@field y number
---@operator add(Vec2|number): Vec2
---@operator sub(Vec2|number): Vec2
---@operator mul(Vec2|number): Vec2
---@operator div(number): Vec2
---@operator unm: Vec2
Vec2 = {
    __add = function(a, b)
        if type(b) == "number" then
            return Vec2:new(a.x + b, a.x + b)
        elseif type(a) == "number" then
            return Vec2:new(a + b.x, a + b.y)
        else
            return Vec2:new(a.x + b.x, a.y + b.y)
        end
    end,
    __sub = function(a, b)
        if type(b) == "number" then
            return Vec2:new(a.x - b, a.x - b)
        elseif type(a) == "number" then
            return Vec2:new(a - b.x, a - b.y)
        else
            return Vec2:new(a.x - b.x, a.y - b.y)
        end
    end,
    __mul = function(a, b)
        if type(a) == "number" then
            return Vec2:new(a * b.x, a * b.y)
        else
            return Vec2:new(a.x * b, a.y * b)
        end
    end,
    __div = function(a, b)
        return Vec2:new(a.x / b, a.y / b)
    end,
    __unm = function(a)
        return Vec2:new(-a.x, -a.y)
    end,
    __eq = function(a, b)
        return a.x == b.x and a.y == b.y
    end,
    __newindex = function(a, b, c)
        error("Vec2 cannot be mutated")
    end,
    __tostring = function(a)
        return "Vec2{" .. a.x .. ", " .. a.y .. "}"
    end,
}

--- Creates a new vector, with two values. The parameters `x` and `y` are
--- used to represent a point/vector of the form `(x,y)`
---
--- Examples:
---  - `myVector = Vector:new(3, 4)` creates a new vector, `(3, 4)`.
---  - `myVector.x` is `3`.
---  - `myVector.y` is `4`.
---@param x number
---@param y number
---@return Vec2
function Vec2:new(x, y)
    local v = {
        x = x,
        y = y,
    }
    setmetatable(v, self)
    self.__index = self

    return v
end

--- Returns the length of the vector.
---
--- Examples:
---  - `myVector = Vector:new(3, 4)` creates a new vector, `(3, 4)`.
---  - `myVector:length()` is `5.0`.
---@return number Length
function Vec2:length()
    return math.sqrt(self.x * self.x + self.y * self.y)
end

--- Returns the vector, except scaled so that its length is 1
---
--- Examples:
---  - `myVector = Vector:new(3, 4)` creates a new vector, `(3, 4)`.
---  - `myVector:normalized()` returns a new vector, `(0.6, 0.8)`.
---  - `myVector:normalized():length()` will always be 1.
---@return Vec2
function Vec2:normalized()
    return self / self:length()
end

--- Returns the vector rotated `radAng` radians
---
--- Examples:
---  - `myVector = Vector:new(3, 4)` creates a new vector, `(3, 4)`.
---  - `myVector:rotate(math.rad(180))` returns a new vector, `(-3, -4)`.
---@param radAng number
---@return Vec2
function Vec2:rotate(radAng)
    return Vec2:new(
        (self.x * math.cos(radAng)) - (self.y * math.sin(radAng)),
        (self.x * math.sin(radAng)) + (self.y * math.cos(radAng))
    )
end

---@param vec Vec2
---@return number
function Vec2:dot(vec)
    return self.x * vec.x + self.y * vec.y
end
