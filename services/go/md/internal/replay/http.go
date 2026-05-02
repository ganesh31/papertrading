package replay

import (
	"encoding/json"
	"log"
	"net/http"
)

type statusWire struct {
	Running      bool    `json:"running"`
	VirtualTime  string  `json:"virtualTime"`
	Speed        float64 `json:"speed"`
	TicksEmitted uint64  `json:"ticksEmitted"`
	SessionID    string  `json:"sessionId"`
	Error        string  `json:"error"`
}

// RegisterHTTP mounts replay control endpoints on mux.
func (c *Coordinator) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("GET /replay/status", c.handleStatus)
	mux.HandleFunc("POST /replay/start", c.handleStart)
	mux.HandleFunc("POST /replay/stop", c.handleStop)
}

func (c *Coordinator) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	st := c.Snapshot()
	wire := statusWire{
		Running:      st.Running,
		Speed:        st.Speed,
		TicksEmitted: st.TicksEmitted,
		SessionID:    st.SessionID,
		Error:        st.Error,
	}
	if !st.VirtualTime.IsZero() {
		wire.VirtualTime = st.VirtualTime.UTC().Format(timeRFC3339Milli)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wire)
}

const timeRFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

func (c *Coordinator) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if c.pool == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "replay: DATABASE_URL not configured",
		})
		return
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := c.Start(req); err != nil {
		log.Printf("replay start: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (c *Coordinator) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c.Stop()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
