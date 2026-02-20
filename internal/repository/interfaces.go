package repository

import "hackaton-service-worker/internal/entity"

type VideoRepository interface {
	FindByID(id string) (*entity.Video, error)
	Update(video *entity.Video) error
}