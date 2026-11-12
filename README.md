# Transcodarr
Automate transcoding media using handbrake

## Run
```
# Using software transcoding
docker run -p 8080:8080 -v ./config:/app/config -v /path/on/host:/path/in/container ghcr.io/alivehamster/transcodarr:latest

# Intel gpu (QSV)
docker run -p 8080:8080 -v ./config:/app/config -v /path/on/host:/path/in/container /dev/dri:/dev/dri ghcr.io/alivehamster/transcodarr:latest


# nvidia gpu (NVENC/NVDEC)
# requires NVIDIA Container Toolkit
docker run -p 8080:8080 -v ./config:/app/config -v /path/on/host:/path/in/container --gpus all ghcr.io/alivehamster/transcodarr:latest
```

## Filters
- Skiplist
- File age
- Hardlink count
- Current media codec
- Minimum file size
- Transcoded media filesize compared to original

## Todo
- Allow inverting filters
- Group skips by directory
- Fix History page
- Time since last scan
- Webhooks after scans