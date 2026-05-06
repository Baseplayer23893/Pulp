package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/config"
	"github.com/Baseplayer23893/Pulp/internal/defuddle"
	"github.com/spf13/cobra"
)

var doctorNetworkURL string

var doctorCmd = &cobra.Command{
	Use:          "doctor",
	Short:        "Check Pulp dependencies and local setup",
	SilenceUsage: true,
	Long: `Check whether Pulp can run its core local workflows.

Doctor verifies extractor dependencies, clipboard tools, config/output
permissions, and a small network fetch.`,
	RunE: runDoctor,
}

type doctorCheck struct {
	Name   string
	OK     bool
	Detail string
	Fix    string
}

func init() {
	doctorCmd.Flags().StringVar(&doctorNetworkURL, "network-url", "https://example.com", "URL to fetch for the network check")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	checks := []doctorCheck{
		checkDefuddle(),
		checkYTDLP(exec.LookPath),
		checkClipboardTools(runtime.GOOS, os.Getenv, exec.LookPath),
		checkConfigWritable(),
		checkOutputWritable(),
		checkNetworkFetch(doctorNetworkURL, 5*time.Second),
	}

	failures := 0
	for _, check := range checks {
		status := "PASS"
		if !check.OK {
			status = "FAIL"
			failures++
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", status, check.Name, check.Detail)
		if !check.OK && check.Fix != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "      Fix: %s\n", check.Fix)
		}
	}

	if failures > 0 {
		return fmt.Errorf("doctor found %d issue(s)", failures)
	}
	return nil
}

func checkDefuddle() doctorCheck {
	if defuddle.IsInstalled() {
		return doctorCheck{Name: "defuddle", OK: true, Detail: "installed"}
	}
	return doctorCheck{
		Name:   "defuddle",
		Detail: "not found",
		Fix:    "install with: npm install -g defuddle",
	}
}

func checkYTDLP(lookPath func(string) (string, error)) doctorCheck {
	if path, err := lookPath("yt-dlp"); err == nil {
		return doctorCheck{Name: "yt-dlp", OK: true, Detail: path}
	}
	return doctorCheck{
		Name:   "yt-dlp",
		Detail: "not found",
		Fix:    "install with: pipx install yt-dlp",
	}
}

func checkClipboardTools(goos string, getenv func(string) string, lookPath func(string) (string, error)) doctorCheck {
	switch goos {
	case "linux", "freebsd", "netbsd", "openbsd", "dragonfly", "solaris":
		if getenv("WAYLAND_DISPLAY") != "" && commandPairExists(lookPath, "wl-copy", "wl-paste") {
			return doctorCheck{Name: "clipboard", OK: true, Detail: "wl-clipboard available"}
		}
		if commandPairExists(lookPath, "xclip", "xclip") {
			return doctorCheck{Name: "clipboard", OK: true, Detail: "xclip available"}
		}
		if commandPairExists(lookPath, "xsel", "xsel") {
			return doctorCheck{Name: "clipboard", OK: true, Detail: "xsel available"}
		}
		return doctorCheck{
			Name:   "clipboard",
			Detail: "no supported clipboard tools found",
			Fix:    "install wl-clipboard on Wayland, or xclip/xsel on X11",
		}
	case "darwin":
		if commandPairExists(lookPath, "pbcopy", "pbpaste") {
			return doctorCheck{Name: "clipboard", OK: true, Detail: "pbcopy/pbpaste available"}
		}
		return doctorCheck{Name: "clipboard", Detail: "pbcopy/pbpaste not found"}
	case "windows":
		if commandPairExists(lookPath, "clip.exe", "powershell.exe") {
			return doctorCheck{Name: "clipboard", OK: true, Detail: "Windows clipboard tools available"}
		}
		return doctorCheck{Name: "clipboard", Detail: "Windows clipboard tools not found"}
	default:
		return doctorCheck{Name: "clipboard", Detail: "unsupported OS: " + goos}
	}
}

func commandPairExists(lookPath func(string) (string, error), first, second string) bool {
	if _, err := lookPath(first); err != nil {
		return false
	}
	_, err := lookPath(second)
	return err == nil
}

func checkConfigWritable() doctorCheck {
	if err := config.EnsureConfigDir(); err != nil {
		return doctorCheck{
			Name:   "config directory",
			Detail: err.Error(),
			Fix:    "ensure the config parent directory is writable",
		}
	}
	return checkDirectoryWritable("config directory", config.ConfigDir())
}

func checkOutputWritable() doctorCheck {
	outDir := config.ResolveOutputDir("")
	if outDir == "" || outDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return doctorCheck{Name: "output directory", Detail: err.Error()}
		}
		outDir = wd
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return doctorCheck{
			Name:   "output directory",
			Detail: err.Error(),
			Fix:    "update output_dir with: pulp config set output_dir <path>",
		}
	}
	return checkDirectoryWritable("output directory", outDir)
}

func checkDirectoryWritable(name, dir string) doctorCheck {
	tmp, err := os.CreateTemp(dir, ".pulp-doctor-*")
	if err != nil {
		return doctorCheck{
			Name:   name,
			Detail: dir + " is not writable: " + err.Error(),
			Fix:    "check directory ownership and permissions",
		}
	}
	path := tmp.Name()
	closeErr := tmp.Close()
	removeErr := os.Remove(path)
	if closeErr != nil {
		return doctorCheck{Name: name, Detail: closeErr.Error()}
	}
	if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return doctorCheck{Name: name, Detail: removeErr.Error()}
	}
	return doctorCheck{Name: name, OK: true, Detail: dir + " is writable"}
}

func checkNetworkFetch(rawURL string, timeout time.Duration) doctorCheck {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return doctorCheck{Name: "network", Detail: "network URL is empty"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return doctorCheck{Name: "network", Detail: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return doctorCheck{
			Name:   "network",
			Detail: err.Error(),
			Fix:    "check internet access, DNS, proxy, or firewall settings",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return doctorCheck{
			Name:   "network",
			Detail: fmt.Sprintf("%s returned HTTP %d", rawURL, resp.StatusCode),
			Fix:    "try a known reachable URL with: pulp doctor --network-url <url>",
		}
	}
	return doctorCheck{Name: "network", OK: true, Detail: rawURL + " reachable"}
}
