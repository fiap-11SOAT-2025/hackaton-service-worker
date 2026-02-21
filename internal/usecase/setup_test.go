package usecase_test

import (
	"github.com/stretchr/testify/mock"
	"hackaton-service-worker/internal/entity"
)

type MockVideoRepository struct{ mock.Mock }

func (m *MockVideoRepository) FindByID(id string) (*entity.Video, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Video), args.Error(1)
}
func (m *MockVideoRepository) Update(v *entity.Video) error { return m.Called(v).Error(0) }

type MockFileStorage struct{ mock.Mock }

func (m *MockFileStorage) Download(b, k, d string) error { return m.Called(b, k, d).Error(0) }
func (m *MockFileStorage) Upload(b, k, s string) error   { return m.Called(b, k, s).Error(0) }

type MockMediaProcessor struct{ mock.Mock }

func (m *MockMediaProcessor) ExtractFrames(v, o string) error { return m.Called(v, o).Error(0) }
func (m *MockMediaProcessor) ZipDirectory(s, z string) error  { return m.Called(s, z).Error(0) }

type MockNotifier struct{ mock.Mock }

// SendNotification implements [usecase.Notifier].
func (m *MockNotifier) SendNotification(email string, videoID string, status string) error {
	panic("unimplemented")
}

func (m *MockNotifier) NotifyError(id, email, msg string) error {
	return m.Called(id, email, msg).Error(0)
}
