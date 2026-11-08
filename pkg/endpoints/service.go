package endpoints

import (
	"context"
	"net/url"
)

// HTTPRequester abstracts the underlying HTTP transport.
type HTTPRequester interface {
	Get(ctx context.Context, path string, query url.Values, target any) error
}

// Service defines operations for managing and querying Portainer environments/endpoints.
type Service interface {
	List(ctx context.Context) ([]Endpoint, error)
	GetMap(ctx context.Context) (map[int]string, error)
}

type serviceImpl struct {
	client HTTPRequester
}

// NewService creates a new Endpoints service instance.
func NewService(client HTTPRequester) Service {
	return &serviceImpl{client: client}
}

func (s *serviceImpl) List(ctx context.Context) ([]Endpoint, error) {
	var list []Endpoint
	if err := s.client.Get(ctx, "/endpoints", nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *serviceImpl) GetMap(ctx context.Context) (map[int]string, error) {
	list, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(list))
	for _, ep := range list {
		m[ep.ID] = ep.Name
	}
	return m, nil
}
