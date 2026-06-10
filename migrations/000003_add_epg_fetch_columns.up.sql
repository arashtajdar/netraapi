ALTER TABLE live_tv_channels 
ADD COLUMN epg_fetch_url VARCHAR(255) NULL,
ADD COLUMN last_epg_fetch DATETIME NULL,
ADD COLUMN next_epg_fetch DATETIME NULL;
