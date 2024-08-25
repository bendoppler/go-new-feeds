local json = require "json"
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"


local users = {}
local user_count = 0

-- Function to load users from CSV file
function load_users_from_csv(filename)
    local file = io.open(filename, "r")
    for line in file:lines() do
        -- Skip the header line
        line = line:gsub("\r", "")
        if line:find("username,password") == nil then
            local username, password = line:match("([^,]+),([^,]+)")
            table.insert(users, {username = username, password = password})
            user_count = user_count + 1
        end
    end
    file:close()
end

-- Load users from the CSV file
load_users_from_csv("users.csv")

function request()
    -- Pick a random user
    local user = users[math.random(1, user_count)]
    -- Create a table for the JSON body
    local body_table = {
        user_name = user.username,
        password = user.password
    }
    -- Convert the table to a JSON string
    wrk.body = json.encode(body_table)
    return wrk.format(nil, "/v1/users/login")
end
