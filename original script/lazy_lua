local utils = require 'mp.utils'
local has_loaded = false
local playlist_titles = {} 
local stream_to_index = {} 
local current_playing_index = 1
local stored_radio_url = nil

local function draw_ui()
    io.stdout:write("\27[H\27[J") 
    for i, title in ipairs(playlist_titles) do
        if i == current_playing_index then
            io.stdout:write("\27[1;35m+ " .. title .. " (playing)\27[0m\n")
        else
            io.stdout:write("+ " .. title .. "\n")
        end
    end
    io.stdout:write("\n")
    io.stdout:flush()
end

-- Sequential extraction remains the same
local function fetch_raw_url(ids_to_fetch, idx)
    if idx > #ids_to_fetch then return end
    local video_url = "https://www.youtube.com/watch?v=" .. ids_to_fetch[idx]
    local command = {
        name = "subprocess",
        capture_stdout = true,
        args = {
            "yt-dlp", "--print", "%(title)s<SEP>%(url)s",
            "--format", "251/140/bestaudio/best",
            "--extractor-args", "youtube:player_client=android;skip=webpage",
            "--force-ipv4", "--no-warnings", "--no-check-certificates",
            video_url
        }
    }
    mp.command_native_async(command, function(success, result)
        if success and result.status == 0 and result.stdout then
            local title, raw_url = string.match(result.stdout, "(.-)<SEP>(https?://%S+)")
            if title and raw_url then
                table.insert(playlist_titles, title)
                stream_to_index[raw_url] = #playlist_titles
                mp.commandv("loadfile", raw_url, "append-play")
                draw_ui()
            end
        end
        fetch_raw_url(ids_to_fetch, idx + 1)
    end)
end

-- Fetch IDs now takes a range so we can request "32-52", "53-73", etc.
local function fetch_ids(radio_url, start_idx, end_idx)
    local command = {
        name = "subprocess",
        capture_stdout = true,
        args = {
            "yt-dlp", "--flat-playlist", 
            "--playlist-items", start_idx .. "-" .. end_idx, 
            "--print", "id", "--no-warnings", "--no-check-certificates",
            radio_url
        }
    }
    mp.command_native_async(command, function(success, result)
        if success and result.status == 0 and result.stdout then
            local ids = {}
            for id in string.gmatch(result.stdout, "[^\r\n]+") do table.insert(ids, id) end
            fetch_raw_url(ids, 1)
        end
    end)
end

mp.register_event("file-loaded", function()
    local opts = mp.get_property("options/script-opts")
    local radio_url = opts and opts:match("radio_url=([^,]+)")
    local first_title = opts and opts:match("first_title=([^,]+)")

    if radio_url and not has_loaded then
        has_loaded = true
        stored_radio_url = radio_url
        table.insert(playlist_titles, first_title or "Initial Track")
        local path = mp.get_property("path")
        if path then stream_to_index[path] = 1 end
        draw_ui()
        -- Start with the first 20 tracks
        fetch_ids(radio_url, 2, 21)
    end
end)

mp.register_event("start-file", function()
    local path = mp.get_property("path")
    if stream_to_index[path] then
        current_playing_index = stream_to_index[path]
        draw_ui()
    end

    -- INFINITE CHECK: If we are near the end of the loaded titles, fetch 20 more
    local total_loaded = #playlist_titles
    if current_playing_index > (total_loaded - 5) and stored_radio_url then
        -- Request the next block from the YouTube Mix
        fetch_ids(stored_radio_url, total_loaded + 1, total_loaded + 21)
    end
end)