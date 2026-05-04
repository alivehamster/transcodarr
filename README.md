# Transcodarr
Automanically transcode media in a folder using handbrake

## Filters
- Skiplist
- File age
- Hardlink count
- Current media codec
- Transcoded media filesize compared to original

## Todo
- Switch skiplist from json in db to its own table with foreign key to a library
- Webpage for history
- Add or remove handbrake configs 
- Manually trigger scan
- Time since last scan
- Time till next scan
- Manually add items to skiplist
- Webhooks after scans