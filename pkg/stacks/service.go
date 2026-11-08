package stacks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// HTTPRequester abstracts the underlying HTTP transport for the stacks service.
type HTTPRequester interface {
	Get(ctx context.Context, path string, query url.Values, target any) error
	PostJSON(ctx context.Context, path string, query url.Values, payload any, target any) error
	PutJSON(ctx context.Context, path string, query url.Values, payload any, target any) error
	Delete(ctx context.Context, path string, query url.Values) error
	PostMultipart(ctx context.Context, path string, query url.Values, formFields map[string]string, fileFieldName string, filePath string, target any) error
}

// Service defines operations for managing Portainer stacks.
type Service interface {
	List(ctx context.Context, filter *StackFilter) ([]Stack, error)
	Inspect(ctx context.Context, id int) (*Stack, error)
	GetFile(ctx context.Context, id int) (string, error)
	Start(ctx context.Context, id int, endpointID int) (*Stack, error)
	Stop(ctx context.Context, id int, endpointID int) (*Stack, error)
	Restart(ctx context.Context, id int, endpointID int) (*Stack, error)
	Delete(ctx context.Context, id int, endpointID int, external bool) error
	CreateComposeFromString(ctx context.Context, endpointID int, payload *CreateComposeStringPayload) (*Stack, error)
	CreateComposeFromFile(ctx context.Context, endpointID int, name string, filePath string, env []Pair) (*Stack, error)
	CreateComposeFromGit(ctx context.Context, endpointID int, payload *CreateComposeGitPayload) (*Stack, error)
	CreateSwarmFromString(ctx context.Context, endpointID int, payload *CreateSwarmStringPayload) (*Stack, error)
	CreateSwarmFromGit(ctx context.Context, endpointID int, payload *CreateSwarmGitPayload) (*Stack, error)
	Update(ctx context.Context, id int, endpointID int, payload *UpdateStackPayload) (*Stack, error)
	RedeployGit(ctx context.Context, id int, endpointID int, payload *RedeployGitPayload) (*Stack, error)
	Migrate(ctx context.Context, id int, endpointID int, payload *MigrateStackPayload) (*Stack, error)
	Associate(ctx context.Context, id int, endpointID int, orphanSwarmID string) (*Stack, error)
}

type serviceImpl struct {
	client HTTPRequester
}

// NewService creates a new Stacks service instance.
func NewService(client HTTPRequester) Service {
	return &serviceImpl{client: client}
}

func (s *serviceImpl) List(ctx context.Context, filter *StackFilter) ([]Stack, error) {
	q := url.Values{}
	if filter != nil {
		fMap := make(map[string]any)
		if filter.EndpointID > 0 {
			fMap["EndpointID"] = filter.EndpointID
		}
		if filter.SwarmID != "" {
			fMap["SwarmID"] = filter.SwarmID
		}
		if len(fMap) > 0 {
			rawJSON, err := json.Marshal(fMap)
			if err != nil {
				return nil, fmt.Errorf("failed to encode stack filters: %w", err)
			}
			q.Set("filters", string(rawJSON))
		}
	}

	var stacksList []Stack
	if err := s.client.Get(ctx, "/stacks", q, &stacksList); err != nil {
		return nil, err
	}
	return stacksList, nil
}

func (s *serviceImpl) Inspect(ctx context.Context, id int) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d", id)
	var stack Stack
	if err := s.client.Get(ctx, path, nil, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) GetFile(ctx context.Context, id int) (string, error) {
	path := fmt.Sprintf("/stacks/%d/file", id)
	var resp StackFileResponse
	if err := s.client.Get(ctx, path, nil, &resp); err != nil {
		return "", err
	}
	return resp.StackFileContent, nil
}

func (s *serviceImpl) Start(ctx context.Context, id int, endpointID int) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d/start", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, nil, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) Stop(ctx context.Context, id int, endpointID int) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d/stop", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, nil, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) Restart(ctx context.Context, id int, endpointID int) (*Stack, error) {
	current, err := s.Inspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed inspecting stack before restart: %w", err)
	}

	if endpointID <= 0 {
		endpointID = current.EndpointID
	}

	// If currently active, stop the stack first
	if current.Status == StackStatusActive {
		if _, err := s.Stop(ctx, id, endpointID); err != nil {
			return nil, fmt.Errorf("failed stopping stack during restart: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	// Start the stack
	started, err := s.Start(ctx, id, endpointID)
	if err != nil {
		return nil, fmt.Errorf("failed starting stack during restart: %w", err)
	}

	return started, nil
}

func (s *serviceImpl) Delete(ctx context.Context, id int, endpointID int, external bool) error {
	path := fmt.Sprintf("/stacks/%d", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	if external {
		q.Set("external", "true")
	}
	return s.client.Delete(ctx, path, q)
}

func (s *serviceImpl) CreateComposeFromString(ctx context.Context, endpointID int, payload *CreateComposeStringPayload) (*Stack, error) {
	path := "/stacks/create/standalone/string"
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) CreateComposeFromFile(ctx context.Context, endpointID int, name string, filePath string, env []Pair) (*Stack, error) {
	path := "/stacks/create/standalone/file"
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}

	fields := map[string]string{
		"Name": name,
	}
	if len(env) > 0 {
		envJSON, err := json.Marshal(env)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling env vars: %w", err)
		}
		fields["Env"] = string(envJSON)
	}

	var stack Stack
	if err := s.client.PostMultipart(ctx, path, q, fields, "file", filePath, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) CreateComposeFromGit(ctx context.Context, endpointID int, payload *CreateComposeGitPayload) (*Stack, error) {
	path := "/stacks/create/standalone/repository"
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) CreateSwarmFromString(ctx context.Context, endpointID int, payload *CreateSwarmStringPayload) (*Stack, error) {
	path := "/stacks/create/swarm/string"
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) CreateSwarmFromGit(ctx context.Context, endpointID int, payload *CreateSwarmGitPayload) (*Stack, error) {
	path := "/stacks/create/swarm/repository"
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) Update(ctx context.Context, id int, endpointID int, payload *UpdateStackPayload) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PutJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) RedeployGit(ctx context.Context, id int, endpointID int, payload *RedeployGitPayload) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d/git/redeploy", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PutJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) Migrate(ctx context.Context, id int, endpointID int, payload *MigrateStackPayload) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d/migrate", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	var stack Stack
	if err := s.client.PostJSON(ctx, path, q, payload, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *serviceImpl) Associate(ctx context.Context, id int, endpointID int, orphanSwarmID string) (*Stack, error) {
	path := fmt.Sprintf("/stacks/%d/associate", id)
	q := url.Values{}
	if endpointID > 0 {
		q.Set("endpointId", strconv.Itoa(endpointID))
	}
	if orphanSwarmID != "" {
		q.Set("orphanedSwarmId", orphanSwarmID)
	}
	var stack Stack
	if err := s.client.PutJSON(ctx, path, q, nil, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}
