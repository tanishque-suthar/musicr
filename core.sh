#!/usr/bin/env bash
# Load configuration
load_config() {
    local config_file="${CONFIG_DIR}/config"
    if [ -f "$config_file" ]; then
        source "$config_file"
    fi
}

# Extract video ID, title, and stream URL from yt-dlp
get_video_info() {
    local search_query="$1"
    
    local output=$(yt-dlp --print id --print title --print urls \
        --format "${VIDEO_FORMAT:-251/140/bestaudio/best}" \
        --extractor-args "youtube:player_client=${PLAYER_CLIENT:-android};skip=webpage" \
        --force-ipv4 --no-warnings "ytsearch1:$search_query" 2>/dev/null)
    
    echo "$output"
}

# Clean title for display
clean_title() {
    local title="$1"
    echo "$title" | tr -d ',='
}

# Play music with mpv and lazy-radio script
play_music() {
    local search_query="$1"
    
    load_config
    
    # Get video info
    local output=$(get_video_info "$search_query")
    local id=$(echo "$output" | sed -n '1p')
    local title=$(echo "$output" | sed -n '2p')
    local stream_url=$(echo "$output" | sed -n '3p')
    
    if [ -z "$id" ] || [ -z "$stream_url" ]; then
        echo "Error: Could not find video" >&2
        return 1
    fi
    
    local clean_name=$(clean_title "$title")
    
    # Get the Lua script path
    local lua_script="$CONFIG_DIR/scripts/lazy-radio.lua"
    
    # Build mpv command with options from config
    mpv \
        --no-video \
        ${TERM_STATUS_MSG:+--term-status-msg="$TERM_STATUS_MSG"} \
        --script="$lua_script" \
        --script-opts="radio_url=https://www.youtube.com/watch?v=$id&list=RD$id,first_title=$clean_name" \
        --ytdl-raw-options="ignore-config=,no-warnings=,no-check-certificates=,format=${VIDEO_FORMAT:-251/140/bestaudio/best},force-ipv4=,extractor-args=youtube:player_client=${PLAYER_CLIENT:-android}" \
        --cache=yes \
        --demuxer-max-bytes=${DEMUXER_MAX_BYTES:-67MiB} \
        --prefetch-playlist=yes \
        "$stream_url"
}

# Check if dependencies are installed
check_deps() {
    local deps=(yt-dlp mpv)
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing dependencies: ${missing[*]}" >&2
        return 1
    fi
}

export -f load_config
export -f get_video_info
export -f clean_title
export -f play_music
export -f check_deps