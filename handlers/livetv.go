package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

// LiveTVHandler handles HTTP requests for the Live TV domain.
type LiveTVHandler struct {
	livetvService *services.LiveTVService
}

// NewLiveTVHandler creates a new LiveTVHandler.
func NewLiveTVHandler(livetvService *services.LiveTVService) *LiveTVHandler {
	return &LiveTVHandler{livetvService: livetvService}
}

func (h *LiveTVHandler) GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	channels, err := h.livetvService.ListChannels(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(channels)
}

// YoutubeEmbed serves a YouTube player html with proper referrer policies to bypass Error 153.
func (h *LiveTVHandler) YoutubeEmbed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YouTube Player</title>
    <style>
        body, html { 
            margin: 0; 
            padding: 0; 
            width: 100%%; 
            height: 100%%; 
            overflow: hidden; 
            background-color: #000; 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            user-select: none;
            -webkit-user-select: none;
        }
        
        .video-wrapper {
            position: relative;
            width: 100%%;
            height: 100%%;
            background-color: #000;
            overflow: hidden;
        }

        iframe { 
            width: 100%%; 
            height: 100%%; 
            border: none; 
            display: block;
        }

        .video-overlay {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            z-index: 10;
            background: transparent;
            pointer-events: auto;
            transition: background-color 0.3s ease;
        }

        .video-overlay.unlocked {
            pointer-events: none;
        }

        /* Floating premium control panel */
        .video-controls {
            position: absolute;
            top: 15px;
            right: 15px;
            z-index: 20;
            display: flex;
            gap: 8px;
            background: rgba(15, 23, 42, 0.75);
            backdrop-filter: blur(12px);
            -webkit-backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.12);
            padding: 6px;
            border-radius: 12px;
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.5);
            opacity: 0.85;
            transition: opacity 0.3s ease, transform 0.3s ease;
            pointer-events: auto;
        }

        .video-controls:hover {
            opacity: 1;
            transform: translateY(1px);
        }

        .control-btn {
            background: transparent;
            border: none;
            color: rgba(255, 255, 255, 0.8);
            cursor: pointer;
            width: 36px;
            height: 36px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
            padding: 0;
            outline: none;
        }

        .control-btn:hover {
            background: rgba(255, 255, 255, 0.1);
            color: #fff;
        }

        .control-btn:active {
            transform: scale(0.92);
        }

        .control-btn svg {
            width: 20px;
            height: 20px;
            fill: none;
            stroke: currentColor;
            stroke-width: 2;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        /* Active gold styling for locked state */
        .control-btn.locked {
            color: #d4af37;
            background: rgba(212, 175, 55, 0.15);
            border: 1px solid rgba(212, 175, 55, 0.25);
        }

        .control-btn.locked:hover {
            background: rgba(212, 175, 55, 0.25);
            color: #fff;
        }

        /* Tooltip style */
        .control-btn {
            position: relative;
        }

        .control-btn::after {
            content: attr(data-tooltip);
            position: absolute;
            bottom: -35px;
            right: 0;
            background: rgba(15, 23, 42, 0.9);
            color: #fff;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 11px;
            white-space: nowrap;
            opacity: 0;
            pointer-events: none;
            transition: opacity 0.2s ease, transform 0.2s ease;
            transform: translateY(-5px);
            border: 1px solid rgba(255, 255, 255, 0.08);
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        }

        .control-btn:hover::after {
            opacity: 1;
            transform: translateY(0);
        }
    </style>
</head>
<body>
    <div class="video-wrapper">
        <iframe 
            id="yt-player"
            src="https://www.youtube.com/embed/%s?autoplay=1&mute=0&loop=1&playlist=%s&controls=1" 
            referrerpolicy="strict-origin-when-cross-origin"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
            allowfullscreen>
        </iframe>
        <div class="video-overlay" id="video-overlay"></div>
        <div class="video-controls">
            <button id="lock-btn" class="control-btn locked" data-tooltip="Unlock Controls">
                <svg viewBox="0 0 24 24">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
            </button>
            <button id="refresh-btn" class="control-btn" data-tooltip="Refresh Stream">
                <svg viewBox="0 0 24 24">
                    <path d="M23 4v6h-6"></path>
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
                </svg>
            </button>
        </div>
    </div>

    <script>
        const lockBtn = document.getElementById('lock-btn');
        const refreshBtn = document.getElementById('refresh-btn');
        const overlay = document.getElementById('video-overlay');
        const iframe = document.getElementById('yt-player');

        let isLocked = true;

        const lockIconHtml = '<svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>';
        const unlockIconHtml = '<svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 9.9-1"></path></svg>';

        lockBtn.addEventListener('click', () => {
            isLocked = !isLocked;
            if (isLocked) {
                overlay.classList.remove('unlocked');
                lockBtn.classList.add('locked');
                lockBtn.innerHTML = lockIconHtml;
                lockBtn.setAttribute('data-tooltip', 'Unlock Controls');
            } else {
                overlay.classList.add('unlocked');
                lockBtn.classList.remove('locked');
                lockBtn.innerHTML = unlockIconHtml;
                lockBtn.setAttribute('data-tooltip', 'Lock Controls');
            }
        });

        refreshBtn.addEventListener('click', () => {
            // Reloads only the iframe by resetting the src attribute
            const currentSrc = iframe.src;
            iframe.src = '';
            setTimeout(() => {
                iframe.src = currentSrc;
            }, 50);
        });
    </script>
</body>
</html>`, id, id)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
