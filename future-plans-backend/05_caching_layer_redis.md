# 05. Caching Layer (Redis)

## Concept
Right now, every client request hits the MySQL database directly. For a premium, high-traffic streaming platform, this will eventually cause bottlenecks. Integrating a caching layer is essential.

## Implementation Details
1. **Redis Integration:**
   - Add a Redis client library to the Go backend (`github.com/redis/go-redis/v9`).
2. **Key Caching Areas:**
   - **Home Screen Data:** Categories, Featured Items, and static rows that are the same for all users.
   - **EPG Data:** Live TV schedules which are queried constantly but change infrequently.
   - **Session Data:** User JWT validation and active session tracking.
3. **Invalidation Strategy:**
   - When the Admin updates a movie or category in the CMS, automatically purge the relevant Redis keys to keep data fresh.
