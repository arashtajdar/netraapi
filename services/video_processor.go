package services

import (
	"log"
	"os/exec"
	"fmt"
	"path/filepath"
	"strings"
)

type VideoTask struct {
	VideoURL string
	OutputPath string // Where to save the HLS files temporarily before uploading or directly modifying
}

var videoQueue = make(chan VideoTask, 100)

func InitVideoProcessor() {
	go worker()
}

func QueueVideoForProcessing(task VideoTask) {
	select {
	case videoQueue <- task:
		log.Println("Video queued for processing:", task.VideoURL)
	default:
		log.Println("Video queue full, dropping task:", task.VideoURL)
	}
}

func worker() {
	for task := range videoQueue {
		processVideo(task)
	}
}

func processVideo(task VideoTask) {
	log.Println("Starting processing for:", task.VideoURL)
	
	// Create output directory
	baseName := strings.TrimSuffix(filepath.Base(task.VideoURL), filepath.Ext(task.VideoURL))
	outDir := filepath.Join("/tmp", baseName) // Use /tmp for Docker
	
	// Example FFmpeg command for HLS (simplified)
	// In production, you would fetch the file if it's an HTTP URL, or read from disk.
	// For this pipeline, assuming we pass the URL directly to ffmpeg (it supports http inputs)
	
	hlsOutput := filepath.Join(outDir, "master.m3u8")
	
	// Create directory if not exists (using os.MkdirAll but since it's just a shell command simulation we can do it via bash or exec)
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

	// Generate VTT Sprites (Basic example)
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
	
	// TODO: Upload files from outDir back to storage (R2/S3/Local)
	// and update the database with the new m3u8 URL.
}
