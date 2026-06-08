# 06. Video Processing Pipeline (HLS/DASH)

## Concept
Premium platforms serve adaptive streaming formats (changing quality based on internet speed) rather than static MP4 files. This ensures zero buffering for users on slow connections.

## Implementation Details
1. **Background Worker:**
   - Set up a message queue (RabbitMQ) or simple Go background routines.
   - When the admin uploads or registers a new video, queue it for processing.
2. **FFmpeg Integration:**
   - The Go worker will execute FFmpeg commands to transcode the video into 1080p, 720p, 480p versions.
   - Generate an HLS (`.m3u8`) or DASH (`.mpd`) master playlist.
3. **Thumbnail Sprites:**
   - Use FFmpeg to generate VTT sprite sheets so users can see a preview thumbnail when hovering or scrubbing the progress bar.
