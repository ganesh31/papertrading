package ptcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Data ingestion helpers (delegate to infra/seed)",
}

var dataFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Run infra/seed/fetch.py (Python; see docs/phases/phase-01-market-data.md)",
}

var dataFetchMinuteCmd = &cobra.Command{
	Use:          "minute",
	Short:        "1m bars via yfinance → md.bars_1m (wraps infra/seed/fetch.py minute)",
	SilenceUsage: true,
	RunE:         runDataFetchMinute,
}

var dataFetchBhavcopyCmd = &cobra.Command{
	Use:          "bhavcopy",
	Short:        "NSE equity EOD bhav → md.bhav_eq (wraps infra/seed/fetch.py bhavcopy)",
	SilenceUsage: true,
	RunE:         runDataFetchBhavcopy,
}

func init() {
	dataFetchMinuteCmd.Flags().String("symbols", "", "Comma-separated symbols (required), e.g. INFY,RELIANCE")
	_ = dataFetchMinuteCmd.MarkFlagRequired("symbols")
	dataFetchMinuteCmd.Flags().Int("days", 7, "Lookback days (1–7 for yfinance 1m)")
	dataFetchMinuteCmd.Flags().String("database-url", "", "Postgres URL (default: $DATABASE_URL)")

	dataFetchBhavcopyCmd.Flags().String("from", "", "Start date YYYY-MM-DD (required)")
	dataFetchBhavcopyCmd.Flags().String("to", "", "End date YYYY-MM-DD (required)")
	_ = dataFetchBhavcopyCmd.MarkFlagRequired("from")
	_ = dataFetchBhavcopyCmd.MarkFlagRequired("to")
	dataFetchBhavcopyCmd.Flags().String("out-dir", "", "Download dir (default: infra/seed/bhavcopy under repo root)")
	dataFetchBhavcopyCmd.Flags().String("udiff-base", "", "Override PT_NSE_CM_BHAVCOPY_BASE for UDiFF CSV mirror")
	dataFetchBhavcopyCmd.Flags().String("database-url", "", "Postgres URL (default: $DATABASE_URL)")

	dataFetchCmd.AddCommand(dataFetchMinuteCmd)
	dataFetchCmd.AddCommand(dataFetchBhavcopyCmd)
	dataCmd.AddCommand(dataFetchCmd)
	rootCmd.AddCommand(dataCmd)
}

func runDataFetchMinute(cmd *cobra.Command, _ []string) error {
	root, err := RepoRoot()
	if err != nil {
		return err
	}
	symbols, err := cmd.Flags().GetString("symbols")
	if err != nil {
		return err
	}
	days, err := cmd.Flags().GetInt("days")
	if err != nil {
		return err
	}
	dsn, err := cmd.Flags().GetString("database-url")
	if err != nil {
		return err
	}
	args := []string{
		"minute",
		"--symbols=" + symbols,
		fmt.Sprintf("--days=%d", days),
	}
	if dsn != "" {
		args = append(args, "--database-url="+dsn)
	}
	return runFetchPy(cmd.Context(), root, args)
}

func runDataFetchBhavcopy(cmd *cobra.Command, _ []string) error {
	root, err := RepoRoot()
	if err != nil {
		return err
	}
	from, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}
	args := []string{
		"bhavcopy",
		"--from=" + from,
		"--to=" + to,
	}
	if od, _ := cmd.Flags().GetString("out-dir"); od != "" {
		args = append(args, "--out-dir="+od)
	}
	if ub, _ := cmd.Flags().GetString("udiff-base"); ub != "" {
		args = append(args, "--udiff-base="+ub)
	}
	if dsn, _ := cmd.Flags().GetString("database-url"); dsn != "" {
		args = append(args, "--database-url="+dsn)
	}
	return runFetchPy(cmd.Context(), root, args)
}

func runFetchPy(ctx context.Context, repoRoot string, fetchArgs []string) error {
	script := filepath.Join(repoRoot, "infra", "seed", "fetch.py")
	if st, err := os.Stat(script); err != nil || st.IsDir() {
		return fmt.Errorf("missing %s (Phase 1 seed script)", script)
	}
	venvPy := filepath.Join(repoRoot, "infra", "seed", ".venv", "bin", "python3")
	python := "python3"
	if st, err := os.Stat(venvPy); err == nil && !st.IsDir() {
		python = venvPy
	}
	argv := append([]string{script}, fetchArgs...)
	c := exec.CommandContext(ctx, python, argv...)
	c.Dir = repoRoot
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	if err := c.Run(); err != nil {
		return fmt.Errorf("fetch.py: %w", err)
	}
	return nil
}
