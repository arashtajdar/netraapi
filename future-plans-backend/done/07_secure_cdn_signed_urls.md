# 07. Secure CDN & Signed URLs

## Concept
Prevent unauthorized sharing or scraping of your video files. Video files should be served through a CDN, and the actual URLs should expire after a short period.

## Implementation Details
1. **CDN Setup:**
   - Connect a CDN (like Cloudflare, AWS CloudFront, or a custom Nginx edge server) to serve media assets.
2. **Signed URLs Generation (Go):**
   - Refactor the API endpoint that returns `video_sources`.
   - Instead of returning the raw URL, generate a cryptographic hash (using a secret key) and append an expiration timestamp (e.g., valid for 2 hours).
3. **Edge Validation:**
   - Ensure the CDN or edge server validates the signature before serving the video chunk, blocking requests with expired or invalid tokens.
