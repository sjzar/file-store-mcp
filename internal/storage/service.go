package storage

import "context"

type Service struct {
	Storage Storage
}

func NewService() *Service {
	return &Service{
		Storage: InitStorage(),
	}
}

func (s *Service) UploadFile(ctx context.Context, path string) (string, error) {
	return s.Storage.UploadFile(ctx, path)
}
