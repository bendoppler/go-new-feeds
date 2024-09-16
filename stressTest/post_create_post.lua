---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by dothaibao.
--- DateTime: 16/9/24 23:17
---
wrk.method = "POST"
wrk.headers["Content-Type"] = "multipart/form-data; boundary=boundary123"
wrk.headers["Authorization"] = "wrk-stress-test"

-- Generate a multipart body
local text = "This is a sample post"
local imageFileName = "dummy_image.jpg"
local imageContent = "dummy_image_data"  -- Simulating binary data for the image

local body = "--boundary123\r\n" ..
        "Content-Disposition: form-data; name=\"text\"\r\n\r\n" ..
        text .. "\r\n" ..
        "--boundary123\r\n" ..
        "Content-Disposition: form-data; name=\"image\"; filename=\"" .. imageFileName .. "\"\r\n" ..
        "Content-Type: image/jpeg\r\n\r\n" ..
        imageContent .. "\r\n" ..
        "--boundary123--\r\n"

wrk.body = body
