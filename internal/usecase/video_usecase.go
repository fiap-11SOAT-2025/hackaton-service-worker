package usecase

import (
	"hackaton-service-worker/internal/entity"
	"hackaton-service-worker/internal/repository"
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

type Notifier interface {
	NotifyError(videoID, email, errorMsg string) error
	SendNotification(email string, videoID string, status string) error
}

type VideoUseCase struct {
	Repo    repository.VideoRepository
	Storage FileStorage
	Media   MediaProcessor
	Notify  Notifier
}

func NewVideoUseCase(repo repository.VideoRepository, storage FileStorage, media MediaProcessor, notify Notifier) *VideoUseCase {
	return &VideoUseCase{
		Repo:    repo,
		Storage: storage,
		Media:   media,
		Notify:  notify,
	}
}

func (uc *VideoUseCase) Execute(videoID, userEmail string) error {
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
		return uc.handleError(video, userEmail, "Falha no Download: "+err.Error())
	}

	if err := uc.Media.ExtractFrames(localVideo, framesDir); err != nil {
		return uc.handleError(video, userEmail, "Falha na Extração: "+err.Error())
	}

	if err := uc.Media.ZipDirectory(framesDir, zipPath); err != nil {
		return uc.handleError(video, userEmail, "Falha ao Zipar: "+err.Error())
	}

	outputKey := fmt.Sprintf("outputs/%s/images.zip", videoID)
	if err := uc.Storage.Upload(video.InputBucket, outputKey, zipPath); err != nil {
		return uc.handleError(video, userEmail, "Falha no Upload Final: "+err.Error())
	}

	video.Status = entity.StatusDone
	video.OutputBucket = video.InputBucket
	video.OutputKey = outputKey

	log.Printf("✅ Processamento concluído: %s", videoID)
	
	err = uc.Repo.Update(video)
	
	if err == nil && userEmail != "" && uc.Notify != nil {
		go func() {
			uc.Notify.SendNotification(userEmail, videoID, string(entity.StatusDone))
		}()
	}

	return err
}


func (uc *VideoUseCase) handleError(video *entity.Video, email, msg string) error {
	log.Printf("❌ %s", msg)
	video.Status = entity.StatusError
	video.ErrorMessage = msg
	uc.Repo.Update(video)

	if email != "" {
		go func() {
			uc.Notify.NotifyError(video.ID, email, msg)
		}()
	}

	return fmt.Errorf(msg)
}

