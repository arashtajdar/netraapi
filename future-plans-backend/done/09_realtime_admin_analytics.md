# 09. Real-Time Admin Analytics Dashboard

## Concept
A premium streaming platform needs a premium admin panel. Implement data visualization to monitor the health and growth of the platform.

## Implementation Details
1. **Go HTML Templates & Tailwind:**
   - Add new template views in `api/views/` for a dashboard.
   - Use Tailwind CSS to create a beautiful, modern layout with dark mode support.
2. **Charts Integration:**
   - Embed a lightweight JavaScript charting library (like Chart.js or ApexCharts) via CDN.
3. **Metrics to Track:**
   - Concurrent viewers (active sessions).
   - Daily Active Users (DAU) and sign-up trends.
   - Most popular movies/series over the last 7 days.
   - Total virtual coins distributed/spent.
