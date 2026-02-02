package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type MediaService struct{}

func NewMediaService() *MediaService {
	return &MediaService{}
}

func (m *MediaService) ExtractFrames(videoPath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1", filepath.Join(outputDir, "frame_%04d.png"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ffmpeg: %s - %v", string(output), err)
	}
	return nil
}

func (m *MediaService) ZipDirectory(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		w, err := archive.Create(f.Name())
		if err != nil {
			return err
		}
		
		src, err := os.Open(filepath.Join(sourceDir, f.Name()))
		if err != nil {
			return err
		}
		
		if _, err := io.Copy(w, src); err != nil {
			src.Close()
			return err
		}
		src.Close()
	}
	return nil
}