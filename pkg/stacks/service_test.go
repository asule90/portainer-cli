package stacks_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portainer/portainer-cli/pkg/client"
	"github.com/portainer/portainer-cli/pkg/stacks"
)

func TestStacksService_List(t *testing.T) {
	mockStacks := []stacks.Stack{
		{
			ID:         1,
			Name:       "test-stack",
			Type:       stacks.StackTypeStandalone,
			EndpointID: 2,
			Status:     stacks.StackStatusActive,
			CreatedBy:  "admin",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stacks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockStacks)
	}))
	defer server.Close()

	c, err := client.NewClient(
		client.WithBaseURL(server.URL),
		client.WithAPIKey("test-key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	list, err := c.Stacks().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(list))
	}
	if list[0].Name != "test-stack" {
		t.Errorf("expected name 'test-stack', got '%s'", list[0].Name)
	}
	if list[0].Type != stacks.StackTypeStandalone {
		t.Errorf("expected type standalone, got %v", list[0].Type)
	}
}

func TestStacksService_Inspect(t *testing.T) {
	mockStack := stacks.Stack{
		ID:         7,
		Name:       "my-app",
		Type:       stacks.StackTypeStandalone,
		EndpointID: 2,
		Status:     stacks.StackStatusActive,
		EntryPoint: "docker-compose.yml",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stacks/7" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockStack)
	}))
	defer server.Close()

	c, err := client.NewClient(
		client.WithBaseURL(server.URL),
		client.WithAPIKey("test-key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	st, err := c.Stacks().Inspect(context.Background(), 7)
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}

	if st.ID != 7 || st.Name != "my-app" {
		t.Errorf("unexpected stack: %+v", st)
	}
}

func TestStacksService_GetFile(t *testing.T) {
	mockResp := stacks.StackFileResponse{
		StackFileContent: "version: '3'\nservices:\n  app:\n    image: nginx",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stacks/7/file" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	c, err := client.NewClient(
		client.WithBaseURL(server.URL),
		client.WithAPIKey("test-key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	content, err := c.Stacks().GetFile(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetFile returned error: %v", err)
	}

	if content != mockResp.StackFileContent {
		t.Errorf("expected content '%s', got '%s'", mockResp.StackFileContent, content)
	}
}

func TestStacksService_StartStop(t *testing.T) {
	var startCalled, stopCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/stacks/7/start":
			startCalled = true
		case "/api/stacks/7/stop":
			stopCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stacks.Stack{ID: 7, Name: "app"})
	}))
	defer server.Close()

	c, err := client.NewClient(
		client.WithBaseURL(server.URL),
		client.WithAPIKey("test-key"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if _, err := c.Stacks().Start(context.Background(), 7, 2); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !startCalled {
		t.Error("expected start endpoint to be called")
	}

	if _, err := c.Stacks().Stop(context.Background(), 7, 2); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if !stopCalled {
		t.Error("expected stop endpoint to be called")
	}
}
