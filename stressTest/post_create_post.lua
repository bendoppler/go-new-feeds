local json = require "json"

wrk.method = "POST"
wrk.headers["Authorization"] = "wrk-stress-test"
wrk.headers["Content-Type"] = "application/json"

-- List of sample texts for the post
local texts = {
    "Excited to share my latest project!",
    "Just had an amazing meal at a new restaurant!",
    "Can’t believe how quickly the time flies!",
    "Here's a fun fact: Did you know honey never spoils?",
    "What a great day to be productive!",
    "I’m learning something new every day!",
    "Feeling grateful for all the support!",
    "Just finished a great book, highly recommend it!",
    "Here’s a picture from my recent trip!",
    "What's everyone up to this weekend?",
    "I love discovering new music!",
    "Has anyone tried the new café in town?",
    "The weather is perfect for a walk!",
    "Catching up on my favorite TV shows tonight!",
    "I found a great tutorial on Go programming!",
    "Who else is excited for the upcoming movie release?"
}

-- Get the count of texts
local text_count = #texts

-- Seed the random number generator
math.randomseed(os.time())

-- Function to randomly generate post request data
function request()
    -- Pick a random text from the loaded file
    local text = texts[math.random(1, text_count)]

    -- Randomly set hasImage to true or false
    local hasImage = math.random() < 0.5

    -- Create the JSON body for the post
    local body_table = {
        text = text,
        hasImage = hasImage
    }

    -- Convert the table to a JSON string
    wrk.body = json.encode(body_table)

    -- Return the formatted request
    return wrk.format(nil, "/v1/posts")
end
