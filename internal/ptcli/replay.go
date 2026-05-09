package ptcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Start md virtual-clock replay (POST /replay/start); use `pt replay stop` to stop",
	RunE:  runReplayStart,
}

var replayStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "POST /replay/stop on md",
	RunE:  runReplayStop,
}

func init() {
	replayCmd.Flags().String("date", "", "Trading calendar date YYYY-MM-DD (replay bars for this IST day)")
	replayCmd.Flags().String("symbols", "", "Comma-separated symbols e.g. INFY,RELIANCE")
	replayCmd.Flags().Float64("speed", 100, "Replay speed multiplier")
	replayCmd.Flags().Int("ticks-per-bar", 10, "Synthetic ticks per 1m bar")
	replayCmd.Flags().String("session-id", "", "Deterministic replay session id (default from md)")
	replayCmd.Flags().Float64("tick-size", 0, "Tick synth tick size (0 = md default)")
	replayCmd.Flags().Float64("spread-ticks", 0, "Bid/ask spread in ticks (0 = md default)")

	replayCmd.AddCommand(replayStopCmd)
	rootCmd.AddCommand(replayCmd)
}

func mdBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("MD_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:6011"
}

type replayStartBody struct {
	Date         string   `json:"date"`
	Symbols      []string `json:"symbols"`
	Speed        float64  `json:"speed,omitempty"`
	TicksPerBar  int      `json:"ticksPerBar,omitempty"`
	SessionID    string   `json:"sessionId,omitempty"`
	TickSize     float64  `json:"tickSize,omitempty"`
	SpreadTicks  float64  `json:"spreadTicks,omitempty"`
}

func runReplayStart(cmd *cobra.Command, _ []string) error {
	date := strings.TrimSpace(mustString(cmd, "date"))
	if date == "" {
		return fmt.Errorf("--date is required (YYYY-MM-DD)")
	}
	syms := strings.TrimSpace(mustString(cmd, "symbols"))
	if syms == "" {
		return fmt.Errorf("--symbols is required (comma-separated)")
	}
	symbols := splitCSV(syms)
	if len(symbols) == 0 {
		return fmt.Errorf("no symbols after parsing --symbols")
	}
	speed, _ := cmd.Flags().GetFloat64("speed")
	tpb, _ := cmd.Flags().GetInt("ticks-per-bar")
	sess, _ := cmd.Flags().GetString("session-id")
	ts, _ := cmd.Flags().GetFloat64("tick-size")
	st, _ := cmd.Flags().GetFloat64("spread-ticks")

	body := replayStartBody{
		Date:        date,
		Symbols:     symbols,
		Speed:       speed,
		TicksPerBar: tpb,
		SessionID:   strings.TrimSpace(sess),
		TickSize:    ts,
		SpreadTicks: st,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodPost, mdBaseURL()+"/replay/start", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("md replay/start %s: %s", resp.Status, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}

func runReplayStop(_ *cobra.Command, _ []string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, mdBaseURL()+"/replay/stop", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("md replay/stop %s: %s", resp.Status, strings.TrimSpace(string(out)))
	}
	fmt.Println(strings.TrimSpace(string(out)))
	return nil
}

func mustString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToUpper(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
