package stacks

import (
	"fmt"
	"time"
)

// StackType represents the type of a Portainer stack.
// 1 = Docker Swarm stack
// 2 = Docker Compose (Standalone) stack
// 3 = Kubernetes stack
type StackType int

const (
	StackTypeSwarm      StackType = 1
	StackTypeStandalone StackType = 2
	StackTypeKubernetes StackType = 3
)

func (t StackType) String() string {
	switch t {
	case StackTypeSwarm:
		return "swarm"
	case StackTypeStandalone:
		return "compose"
	case StackTypeKubernetes:
		return "kubernetes"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// StackStatus represents the status of a Portainer stack.
// 1 = Active
// 2 = Inactive
type StackStatus int

const (
	StackStatusActive   StackStatus = 1
	StackStatusInactive StackStatus = 2
)

func (s StackStatus) String() string {
	switch s {
	case StackStatusActive:
		return "active"
	case StackStatusInactive:
		return "inactive"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Pair represents an environment variable name-value pair.
type Pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AutoUpdateSettings defines GitOps auto-update configuration for git-backed stacks.
type AutoUpdateSettings struct {
	Interval string `json:"Interval,omitempty"`
	JobID    string `json:"JobId,omitempty"`
	UUID     string `json:"UUID,omitempty"`
	Webhook  string `json:"Webhook,omitempty"`
}

// GitConfig holds git repository configuration for git-deployed stacks.
type GitConfig struct {
	URL                     string `json:"URL,omitempty"`
	ReferenceName           string `json:"ReferenceName,omitempty"`
	ConfigFilePath          string `json:"ConfigFilePath,omitempty"`
	ConfigHash              string `json:"ConfigHash,omitempty"`
	TLSSkipVerify           bool   `json:"TLSSkipVerify,omitempty"`
	Authentication          bool   `json:"Authentication,omitempty"`
	Username                string `json:"Username,omitempty"`
	Password                string `json:"Password,omitempty"`
	GitCredentialID         int    `json:"GitCredentialID,omitempty"`
	ForceUpdate             bool   `json:"ForceUpdate,omitempty"`
}

// ResourceControl defines ownership and access permissions of a stack.
type ResourceControl struct {
	ID                 int      `json:"Id,omitempty"`
	ResourceID         string   `json:"ResourceId,omitempty"`
	SubResourceIDs     []string `json:"SubResourceIds,omitempty"`
	Type               int      `json:"Type,omitempty"`
	Public             bool     `json:"Public,omitempty"`
	AdministratorsOnly bool     `json:"AdministratorsOnly,omitempty"`
	System             bool     `json:"System,omitempty"`
}

// Stack represents a Portainer stack object.
type Stack struct {
	ID              int                 `json:"Id"`
	Name            string              `json:"Name"`
	Type            StackType           `json:"Type"`
	EndpointID      int                 `json:"EndpointId"`
	SwarmID         string              `json:"SwarmId,omitempty"`
	EntryPoint      string              `json:"EntryPoint,omitempty"`
	Env             []Pair              `json:"Env,omitempty"`
	ResourceControl *ResourceControl    `json:"ResourceControl,omitempty"`
	Status          StackStatus         `json:"Status"`
	ProjectPath     string              `json:"ProjectPath,omitempty"`
	CreationDate    int64               `json:"CreationDate,omitempty"`
	CreatedBy       string              `json:"CreatedBy,omitempty"`
	UpdateDate      int64               `json:"UpdateDate,omitempty"`
	UpdatedBy       string              `json:"UpdatedBy,omitempty"`
	AdditionalFiles []string            `json:"AdditionalFiles,omitempty"`
	AutoUpdate      *AutoUpdateSettings `json:"AutoUpdate,omitempty"`
	GitConfig       *GitConfig          `json:"GitConfig,omitempty"`
	FromAppTemplate bool                `json:"FromAppTemplate,omitempty"`
	Namespace       string              `json:"Namespace,omitempty"`
}

// CreatedAt returns CreationDate as a standard time.Time.
func (s *Stack) CreatedAt() time.Time {
	if s.CreationDate == 0 {
		return time.Time{}
	}
	return time.Unix(s.CreationDate, 0)
}

// UpdatedAt returns UpdateDate as a standard time.Time.
func (s *Stack) UpdatedAt() time.Time {
	if s.UpdateDate == 0 {
		return time.Time{}
	}
	return time.Unix(s.UpdateDate, 0)
}

// StackFileResponse represents the response when retrieving a stack's compose/manifest file.
type StackFileResponse struct {
	StackFileContent string `json:"StackFileContent"`
}

// StackFilter represents query filters for StackList.
type StackFilter struct {
	EndpointID int    `json:"EndpointID,omitempty"`
	SwarmID    string `json:"SwarmID,omitempty"`
}

// CreateComposeStringPayload is used for POST /stacks/create/standalone/string.
type CreateComposeStringPayload struct {
	Name             string `json:"Name"`
	StackFileContent string `json:"StackFileContent"`
	Env              []Pair `json:"Env,omitempty"`
	FromAppTemplate  bool   `json:"FromAppTemplate,omitempty"`
}

// CreateComposeGitPayload is used for POST /stacks/create/standalone/repository.
type CreateComposeGitPayload struct {
	Name                    string              `json:"Name"`
	RepositoryURL           string              `json:"RepositoryURL"`
	RepositoryReferenceName string              `json:"RepositoryReferenceName,omitempty"`
	ComposeFile             string              `json:"ComposeFile,omitempty"`
	AdditionalFiles         []string            `json:"AdditionalFiles,omitempty"`
	RepositoryAuthentication bool               `json:"RepositoryAuthentication,omitempty"`
	RepositoryUsername      string              `json:"RepositoryUsername,omitempty"`
	RepositoryPassword      string              `json:"RepositoryPassword,omitempty"`
	TLSSkipVerify           bool                `json:"TLSSkipVerify,omitempty"`
	Env                     []Pair              `json:"Env,omitempty"`
	AutoUpdate              *AutoUpdateSettings `json:"AutoUpdate,omitempty"`
	FromAppTemplate         bool                `json:"FromAppTemplate,omitempty"`
}

// CreateSwarmStringPayload is used for POST /stacks/create/swarm/string.
type CreateSwarmStringPayload struct {
	Name             string `json:"Name"`
	StackFileContent string `json:"StackFileContent"`
	Env              []Pair `json:"Env,omitempty"`
	FromAppTemplate  bool   `json:"FromAppTemplate,omitempty"`
	SwarmID          string `json:"SwarmID,omitempty"`
}

// CreateSwarmGitPayload is used for POST /stacks/create/swarm/repository.
type CreateSwarmGitPayload struct {
	Name                    string              `json:"Name"`
	RepositoryURL           string              `json:"RepositoryURL"`
	RepositoryReferenceName string              `json:"RepositoryReferenceName,omitempty"`
	ComposeFile             string              `json:"ComposeFile,omitempty"`
	AdditionalFiles         []string            `json:"AdditionalFiles,omitempty"`
	RepositoryAuthentication bool               `json:"RepositoryAuthentication,omitempty"`
	RepositoryUsername      string              `json:"RepositoryUsername,omitempty"`
	RepositoryPassword      string              `json:"RepositoryPassword,omitempty"`
	TLSSkipVerify           bool                `json:"TLSSkipVerify,omitempty"`
	Env                     []Pair              `json:"Env,omitempty"`
	AutoUpdate              *AutoUpdateSettings `json:"AutoUpdate,omitempty"`
	FromAppTemplate         bool                `json:"FromAppTemplate,omitempty"`
	SwarmID                 string              `json:"SwarmID,omitempty"`
}

// UpdateStackPayload is used for PUT /stacks/{id}.
type UpdateStackPayload struct {
	StackFileContent       string `json:"StackFileContent,omitempty"`
	Env                    []Pair `json:"Env,omitempty"`
	Prune                  bool   `json:"Prune,omitempty"`
	RepullImageAndRedeploy bool   `json:"RepullImageAndRedeploy,omitempty"`
	PullImage              bool   `json:"PullImage,omitempty"`
}

// RedeployGitPayload is used for PUT /stacks/{id}/git/redeploy.
type RedeployGitPayload struct {
	Env                         []Pair `json:"Env,omitempty"`
	Prune                       bool   `json:"Prune,omitempty"`
	RepullImageAndRedeploy      bool   `json:"RepullImageAndRedeploy,omitempty"`
	RepositoryAuthentication    bool   `json:"RepositoryAuthentication,omitempty"`
	RepositoryUsername          string `json:"RepositoryUsername,omitempty"`
	RepositoryPassword          string `json:"RepositoryPassword,omitempty"`
	RepositoryReferenceName     string `json:"RepositoryReferenceName,omitempty"`
	RepositoryAuthorizationType int    `json:"RepositoryAuthorizationType,omitempty"`
	StackName                   string `json:"StackName,omitempty"`
}

// MigrateStackPayload is used for POST /stacks/{id}/migrate.
type MigrateStackPayload struct {
	EndpointID int    `json:"EndpointID"`
	SwarmID    string `json:"SwarmID,omitempty"`
	Name       string `json:"Name,omitempty"`
}
