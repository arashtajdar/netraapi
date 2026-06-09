# 04. Advanced Recommendation Engine

## Concept
Move beyond simple category mapping. Build a system that analyzes `user_watch_history` to provide personalized rows like "Because you watched X..." or "Top Picks for You".

## Implementation Details
1. **Data Analytics:**
   - Analyze watch history to find overlapping user tastes (Collaborative Filtering) or item metadata similarities (Content-Based Filtering via genres, directors, cast).
2. **Go Backend Logic:**
   - For V1, implement a simplified content-based algorithm: if a user finishes a movie in a specific category, boost the ranking of other highly-rated movies in that category.
   - Create a dedicated `/api/recommendations` endpoint.
3. **Database Performance:**
   - Since these queries can be heavy, results should be pre-calculated daily for active users and cached (e.g., in Redis).
