package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"time"
)

// JSON 
type CommandRequest struct {
	ID		int		`json:"id"`
	Cmd		string	`json:"cmd"`
}

type AgentView struct {
	ID int
	IP string
}

func ui_handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var agentList []AgentView
	for id, agent := range agents {
		agentList = append(agentList, AgentView{ID: id, IP: agent.IP})
	}

	// Parse HTML
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template not found in: /templates/index.html", http.StatusInternalServerError)
		return 
	}

	data := struct {
		Agents []AgentView
	}{
		Agents: agentList,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func listAgentsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var agentList []AgentView
	for id, agent := range agents {
		agentList = append(agentList, AgentView{ID: id, IP: agent.IP})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agentList,
		"count": len(agentList),
	})
}

func sendCommandHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var req CommandRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	mu.Lock()
	agent, exists := agents[req.ID]
	mu.Unlock()

	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	select {
		case agent.InCmdChan <- req.Cmd:
			//Command sent successfully
		case <-time.After(2 * time.Second):
			http.Error(w, "Agent is not ready", http.StatusServiceUnavailable)
			return
	}

	select {
		case output := <-agent.OutCmdChan:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"output": output,
				"status": "success",
			})
		case <-time.After(25 * time.Second):
			http.Error(w, "Agent timeout (25s)", http.StatusGatewayTimeout)
		}
	
}
