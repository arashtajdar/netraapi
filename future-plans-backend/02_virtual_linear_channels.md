# 02. Virtual Linear Channels (24/7 TV) Scheduler

## Concept
Turn VOD content (Movies, Series, Music) into 24/7 "Live" TV channels. This reduces decision fatigue for users and increases watch time by simulating traditional linear television.

## Implementation Details
1. **Cron/Worker Scheduler (Go):**
   - A background worker runs periodically to populate the `epg` (Electronic Program Guide) table.
   - It selects random or curated content based on the channel's theme (e.g., "Classic Movies Channel", "Indie Music Channel").
2. **Database Updates:**
   - Automate the insertion of records into the `epg` table ensuring that `start_time` and `end_time` are perfectly contiguous.
3. **API Endpoints:**
   - Modify the Live TV endpoints so the client can query the current playing VOD file and the exact `seek` time based on `current_time - epg.start_time`.
