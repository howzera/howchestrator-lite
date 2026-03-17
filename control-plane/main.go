package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"howchestrator-lite/shared"
)

// ControlPlane represents the central brain of the orchestrator.
type ControlPlane struct {
	Agents    map[string]shared.AgentRegistration
	Resources map[string]shared.Resource
	mu        sync.RWMutex
}

func NewControlPlane() *ControlPlane {
	return &ControlPlane{
		Agents:    make(map[string]shared.AgentRegistration),
		Resources: make(map[string]shared.Resource),
	}
}

// 📡 HANDLER: Request a new resource (e.g., from a webapp)
func (cp *ControlPlane) handleRequestResource(w http.ResponseWriter, r *http.Request) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Logic: 1. Find an agent with capacity
	//        2. Generate a resource ID and PORT
	//        3. Send a command to the agent to START

	// 1. Pick the first available agent
	var targetAgent shared.AgentRegistration
	found := false
	for _, agent := range cp.Agents {
		// Mock logic: Check current resources vs capacity
		count := 0
		for _, res := range cp.Resources {
			if res.AgentID == agent.AgentID && res.Status != "CLOSED" {
				count++
			}
		}
		if count < agent.Capacity {
			targetAgent = agent
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "No agents available with capacity", http.StatusServiceUnavailable)
		return
	}

	// 2. Mock a Resource
	resourceID := fmt.Sprintf("res-%d", time.Now().UnixNano())
	port := 8000 + len(cp.Resources) // Generic port allocation

	cp.Resources[resourceID] = shared.Resource{
		ID:        resourceID,
		Port:      port,
		Status:    "STARTING",
		AgentID:   targetAgent.AgentID,
		CreatedAt: time.Now(),
	}

	// 3. 🚀 CRITICAL: Command the Agent to START (Async)
	go cp.commandAgent(targetAgent.IP, "START", resourceID, port)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(cp.Resources[resourceID])
	log.Printf("[BRAIN] Resource %s allocated to Agent %s on PORT %d\n", resourceID, targetAgent.AgentID, port)
}

// 📦 HELPER: Send HTTP command to Agent
func (cp *ControlPlane) commandAgent(agentIP string, action string, resID string, port int) {
	cmd := shared.AgentAction{
		Action:     action,
		ResourceID: resID,
		Port:       port,
	}
	body, _ := json.Marshal(cmd)
	url := fmt.Sprintf("http://%s:8081/execute", agentIP) // Agents listen on 8081 in this example

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] Failed to command agent at %s: %v\n", agentIP, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[BRAIN] Command '%s' sent to Agent at %s\n", action, agentIP)
}

// 📥 HANDLER: Webhook from Agent reporting status
func (cp *ControlPlane) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var update shared.ResourceStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	res, ok := cp.Resources[update.ResourceID]
	if !ok {
		log.Printf("[WARN] Received update for unknown resource: %s\n", update.ResourceID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	res.Status = update.Status
	cp.Resources[update.ResourceID] = res

	log.Printf("[WEBHOOK] Node %s reports: Resource %s is now %s (Port: %d)\n", update.AgentID, update.ResourceID, update.Status, update.Port)
	w.WriteHeader(http.StatusOK)
}

// 🤝 HANDLER: Agent Registration
func (cp *ControlPlane) handleRegister(w http.ResponseWriter, r *http.Request) {
	var reg shared.AgentRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cp.mu.Lock()
	cp.Agents[reg.AgentID] = reg
	cp.mu.Unlock()

	log.Printf("[BRAIN] Agent '%s' registered from %s (Capacity: %d)\n", reg.AgentID, reg.IP, reg.Capacity)
	w.WriteHeader(http.StatusOK)
}

func main() {
	cp := NewControlPlane()

	http.HandleFunc("/api/v1/resources", cp.handleRequestResource)
	http.HandleFunc("/api/v1/webhook", cp.handleWebhook)
	http.HandleFunc("/api/v1/register", cp.handleRegister)

	port := ":8080"
	log.Printf("🧠 [CONTROL PLANE] Running on %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
