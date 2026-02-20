package usecase_test

import (
	"errors"
	"hackaton-service-worker/internal/entity"
	"hackaton-service-worker/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVideoUseCase_Execute(t *testing.T) {
	t.Run("Erro: Vídeo não encontrado", func(t *testing.T) {
		repo := new(MockVideoRepository)
		uc := usecase.NewVideoUseCase(repo, nil, nil, nil)
		repo.On("FindByID", "v1").Return(nil, errors.New("not found"))

		err := uc.Execute("v1", "test@test.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "video não encontrado")
	})

	t.Run("Erro: Falha no Download do S3", func(t *testing.T) {
		repo, storage, notify := new(MockVideoRepository), new(MockFileStorage), new(MockNotifier)
		uc := usecase.NewVideoUseCase(repo, storage, nil, notify)

		video := &entity.Video{ID: "v1", InputBucket: "b", InputKey: "k"}
		repo.On("FindByID", "v1").Return(video, nil)
		repo.On("Update", mock.Anything).Return(nil)
		storage.On("Download", "b", "k", mock.Anything).Return(errors.New("s3 fail"))
		
		notify.On("NotifyError", "v1", "user@test.com", mock.Anything).Return(nil)

		err := uc.Execute("v1", "user@test.com")
		assert.Error(t, err)
		assert.Equal(t, entity.StatusError, video.Status)
	})

	t.Run("Erro: Falha na Extração de Frames", func(t *testing.T) {
        repo, storage, media, notify := new(MockVideoRepository), new(MockFileStorage), new(MockMediaProcessor), new(MockNotifier)
        uc := usecase.NewVideoUseCase(repo, storage, media, notify)

        video := &entity.Video{ID: "v1", InputBucket: "b", InputKey: "k"}
        repo.On("FindByID", "v1").Return(video, nil)
        repo.On("Update", mock.Anything).Return(nil)
        storage.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(nil)
        
        media.On("ExtractFrames", mock.Anything, mock.Anything).Return(errors.New("ffmpeg error"))
        notify.On("NotifyError", "v1", "user@test.com", mock.Anything).Return(nil)

        err := uc.Execute("v1", "user@test.com")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "Falha na Extração")
    })

	t.Run("Erro: Falha ao Zipar Diretório", func(t *testing.T) {
        repo, storage, media, notify := new(MockVideoRepository), new(MockFileStorage), new(MockMediaProcessor), new(MockNotifier)
        uc := usecase.NewVideoUseCase(repo, storage, media, notify)

        video := &entity.Video{ID: "v1"}
        repo.On("FindByID", "v1").Return(video, nil)
        repo.On("Update", mock.Anything).Return(nil)
        storage.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(nil)
        media.On("ExtractFrames", mock.Anything, mock.Anything).Return(nil)
        
        media.On("ZipDirectory", mock.Anything, mock.Anything).Return(errors.New("zip error"))
        notify.On("NotifyError", "v1", "user@test.com", mock.Anything).Return(nil)

        err := uc.Execute("v1", "user@test.com")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "Falha ao Zipar")
    })

	t.Run("Erro: Falha no Upload Final", func(t *testing.T) {
        repo, storage, media, notify := new(MockVideoRepository), new(MockFileStorage), new(MockMediaProcessor), new(MockNotifier)
        uc := usecase.NewVideoUseCase(repo, storage, media, notify)

        video := &entity.Video{ID: "v1", InputBucket: "b"}
        repo.On("FindByID", "v1").Return(video, nil)
        repo.On("Update", mock.Anything).Return(nil)
        storage.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(nil)
        media.On("ExtractFrames", mock.Anything, mock.Anything).Return(nil)
        media.On("ZipDirectory", mock.Anything, mock.Anything).Return(nil)
        
        storage.On("Upload", "b", mock.Anything, mock.Anything).Return(errors.New("upload fail"))
        notify.On("NotifyError", "v1", "user@test.com", mock.Anything).Return(nil)

        err := uc.Execute("v1", "user@test.com")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "Falha no Upload Final")
    })

	t.Run("Sucesso: Processamento completo", func(t *testing.T) {
		repo, storage, media := new(MockVideoRepository), new(MockFileStorage), new(MockMediaProcessor)
		uc := usecase.NewVideoUseCase(repo, storage, media, nil)

		video := &entity.Video{ID: "v2", InputBucket: "in", InputKey: "video.mp4"}
		repo.On("FindByID", "v2").Return(video, nil)
		repo.On("Update", mock.Anything).Return(nil)
		
		storage.On("Download", "in", "video.mp4", mock.Anything).Return(nil)
		media.On("ExtractFrames", mock.Anything, mock.Anything).Return(nil)
		media.On("ZipDirectory", mock.Anything, mock.Anything).Return(nil)
		storage.On("Upload", "in", mock.Anything, mock.Anything).Return(nil)

		err := uc.Execute("v2", "user@test.com")
		
		assert.NoError(t, err)
		assert.Equal(t, entity.StatusDone, video.Status)
		assert.NotEmpty(t, video.OutputKey)
	})
}