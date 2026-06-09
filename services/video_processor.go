package services

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sheedbox-api/config"
)

type VideoTask struct {
	VideoURL   string `json:"video_url"`
	OutputPath string `json:"output_path"` // Where to save the HLS files temporarily
}

var videoQueue = make(chan VideoTask, 100)

func InitVideoProcessor() {
	go worker()
}

func QueueVideoForProcessing(task VideoTask) {
	if config.RedisClient != nil {
		data, err := json.Marshal(task)
		if err == nil {
			err = config.RedisClient.LPush(context.Background(), "video_processing_queue", data).Err()
			if err == nil {
				log.Println("Video queued for processing in Redis:", task.VideoURL)
				return
			}
		}
	}

	// Fallback to in-memory channel
	select {
	case videoQueue <- task:
		log.Println("Video queued for processing in fallback channel:", task.VideoURL)
	default:
		log.Println("Video queue full, dropping task:", task.VideoURL)
	}
}

func worker() {
	ctx := context.Background()
	for {
		if config.RedisClient != nil {
			// BRPop blocks up to 5 seconds waiting for a task
			res, err := config.RedisClient.BRPop(ctx, 5*time.Second, "video_processing_queue").Result()
			if err == nil && len(res) == 2 {
				var task VideoTask
				if err := json.Unmarshal([]byte(res[1]), &task); err == nil {
					processVideo(task)
					continue
				}
			}
		}

		// Fallback check on in-memory channel
		select {
		case task := <-videoQueue:
			processVideo(task)
		default:
			// If both Redis and in-memory queue are empty, sleep briefly
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func processVideo(task VideoTask) {
	log.Println("Starting processing for:", task.VideoURL)
	
	// Create output directory
	baseName := strings.TrimSuffix(filepath.Base(task.VideoURL), filepath.Ext(task.VideoURL))
	outDir := filepath.Join("/tmp", baseName) // Use /tmp for Docker
	
	// Example FFmpeg command for HLS (simplified)
	cmdMkdir := exec.Command("mkdir", "-p", outDir)
	cmdMkdir.Run()

	// Transcode to multiple resolutions and create HLS playlist
	cmd := exec.Command("ffmpeg",
		"-i", task.VideoURL,
		"-preset", "veryfast",
		"-g", "48", "-sc_threshold", "0",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:0", "1920x1080", "-c:v:0", "libx264", "-b:v:0", "5000k",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:1", "1280x720", "-c:v:1", "libx264", "-b:v:1", "2800k",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:2", "854x480", "-c:v:2", "libx264", "-b:v:2", "1400k",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",
		filepath.Join(outDir, "stream_%v.m3u8"),
	)
	
	if err := cmd.Run(); err != nil {
		log.Println("Error running FFmpeg for HLS:", err)
		return
	}

	// Generate VTT Sprites
	spriteCmd := exec.Command("ffmpeg",
		"-i", task.VideoURL,
		"-vf", "fps=1/10,scale=160:-1,tile=10x10",
		filepath.Join(outDir, "sprite.jpg"),
	)
	
	if err := spriteCmd.Run(); err != nil {
		log.Println("Error running FFmpeg for Sprites:", err)
		return
	}

	log.Println("Successfully processed video to HLS and Sprites:", outDir)
}
