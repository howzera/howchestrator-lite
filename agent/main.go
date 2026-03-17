package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"howchestrator-lite/shared"
)

// Agent configuration
const (
	AgentID      = "node-1"
	ControlPlane = "http://localhost:8080" // In a real setup, this would be an IP/DNS
	Capacity     = 10
	Port         = ":8081"
)

// Simulated background process state
type ActiveResource struct {
	ID   string
	Port int
}

var ActiveResources = make(map[string]ActiveResource)

// 📥 HANDLER: Receive orders from Control Plane
func handleExecute(w http.ResponseWriter, r *http.Request) {
	var cmd shared.AgentAction
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[AGENT] Command received: %s (Resource: %s, Port: %d)\n", cmd.Action, cmd.ResourceID, cmd.Port)

	// Simulate Resource Creation (Port Opening)
	go func() {
		// ⌛ Start of the "Magic"
		log.Printf("[PROCESS] Setting up environment for %s...\n", cmd.ResourceID)
		time.Sleep(3 * time.Second) // Simulate overhead (Docker Pull, FS Prepare)

		// 🏗️ Step 1: Open Port (Simulated)
		log.Printf("[PROCESS] Port %d is now OPEN for resource %s\n", cmd.Port, cmd.ResourceID)
		ActiveResources[cmd.ResourceID] = ActiveResource{ID: cmd.ResourceID, Port: cmd.Port}

		// 📣 Step 2: Notify Brain via WEBHOOK
		notifyBrain(cmd.ResourceID, cmd.Port, "OPEN", "Resource is live and reachable.")
		
		// ⌛ Keep "Running" for a bit, then simulate closure (Optional)
		// Or wait for a STOP command. For now, let's just stay open.
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "executing"})
}

// 📦 HELPER: Call the Control Plane back
func notifyBrain(resID string, port int, status string, msg string) {
	update := shared.ResourceStatusUpdate{
		AgentID:    AgentID,
		ResourceID: resID,
		Port:       port,
		Status:     status,
		Message:    msg,
	}

	body, _ := json.Marshal(update)
	log.Printf("[WEBHOOK] Sending status '%s' to Brain...\n", status)

	resp, err := http.Post(ControlPlane+"/api/v1/webhook", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[ERROR] Failed to send webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[WEBHOOK] Response from Brain: %d OK\n", resp.StatusCode)
}

// 🤝 HELPER: Register on startup
func register() {
	reg := shared.AgentRegistration{
		AgentID:  AgentID,
		IP:       "localhost", // Mocked
		Capacity: Capacity,
	}
	body, _ := json.Marshal(reg)

	resp, err := http.Post(ControlPlane+"/api/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[INIT ERROR] Brain not found at %s. Waiting...\n", ControlPlane)
		return
	}
	defer resp.Body.Close()
	log.Printf("[INIT] Successfully registered with Brain!\n")
}

func main() {
	// Register with Control Plane
	// In a real agent, this would retry in a loop until success.
	go func() {
		for {
			register()
			time.Sleep(30 * time.Second) // Heartbeat/Re-registration
		}
	}()

	http.HandleFunc("/execute", handleExecute)

	log.Printf("🤖 [AGENT] Node '%s' running on %s...\n", AgentID, Port)
	log.Fatal(http.ListenAndServe(Port, nil))
}
