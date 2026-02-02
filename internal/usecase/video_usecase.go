package usecase

import (
	"fiapx-worker/internal/entity"
	"fiapx-worker/internal/repository"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type FileStorage interface {
	Download(bucket, key, destPath string) error
	Upload(bucket, key, sourcePath string) error
}

type MediaProcessor interface {
	ExtractFrames(videoPath, outputDir string) error
	ZipDirectory(sourceDir, zipPath string) error
}

type VideoUseCase struct {
	Repo    repository.VideoRepository
	Storage FileStorage
	Media   MediaProcessor
}

func NewVideoUseCase(repo repository.VideoRepository, storage FileStorage, media MediaProcessor) *VideoUseCase {
	return &VideoUseCase{
		Repo:    repo,
		Storage: storage,
		Media:   media,
	}
}

func (uc *VideoUseCase) Execute(videoID string) error {
	log.Printf("🔄 Iniciando vídeo: %s", videoID)

	video, err := uc.Repo.FindByID(videoID)
	if err != nil {
		return fmt.Errorf("video não encontrado: %v", err)
	}

	video.Status = entity.StatusProcessing
	uc.Repo.Update(video)

	workDir := filepath.Join("temp", videoID)
	os.MkdirAll(workDir, 0755)
	defer os.RemoveAll(workDir)

	localVideo := filepath.Join(workDir, "input.mp4")
	framesDir := filepath.Join(workDir, "frames")
	zipPath := filepath.Join(workDir, "images.zip")

	if err := uc.Storage.Download(video.InputBucket, video.InputKey, localVideo); err != nil {
		return uc.handleError(video, "Falha no Download: "+err.Error())
	}

	if err := uc.Media.ExtractFrames(localVideo, framesDir); err != nil {
		return uc.handleError(video, "Falha na Extração: "+err.Error())
	}

	if err := uc.Media.ZipDirectory(framesDir, zipPath); err != nil {
		return uc.handleError(video, "Falha ao Zipar: "+err.Error())
	}

	outputKey := fmt.Sprintf("outputs/%s/images.zip", videoID)
	if err := uc.Storage.Upload(video.InputBucket, outputKey, zipPath); err != nil {
		return uc.handleError(video, "Falha no Upload Final: "+err.Error())
	}

	video.Status = entity.StatusDone
	video.OutputBucket = video.InputBucket
	video.OutputKey = outputKey

	log.Printf("✅ Processamento concluído: %s", videoID)
	
	return uc.Repo.Update(video)
}

func (uc *VideoUseCase) handleError(video *entity.Video, msg string) error {
	log.Printf("❌ %s", msg)
	video.Status = entity.StatusError
	video.ErrorMessage = msg
	uc.Repo.Update(video)
	return fmt.Errorf(msg)
}