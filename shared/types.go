package shared

import "time"

// Resource represents a generic "task" or "service" managed by the orchestrator.
type Resource struct {
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	Status    string    `json:"status"` // "STARTING", "OPEN", "CLOSED", "FAILED"
	AgentID   string    `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ResourceRequest is the payload sent to the Control Plane to request a new resource.
type ResourceRequest struct {
	Type     string `json:"type"` // e.g., "game-server", "web-node"
	Priority int    `json:"priority"`
}

// AgentAction is the command sent from the Control Plane to an Agent.
type AgentAction struct {
	Action     string `json:"action"` // "START", "STOP"
	ResourceID string `json:"resource_id"`
	Port       int    `json:"port"`
}

// ResourceStatusUpdate is the webhook payload sent from the Agent back to the Control Plane.
type ResourceStatusUpdate struct {
	AgentID    string `json:"agent_id"`
	ResourceID string `json:"resource_id"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

// AgentRegistration is the heartbeat/registration payload.
type AgentRegistration struct {
	AgentID  string `json:"agent_id"`
	IP       string `json:"ip"`
	Capacity int    `json:"capacity"` // Max concurrent resources
}
