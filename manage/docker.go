package manage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DaisyUIThemes is the complete list of available DaisyUI themes.
var DaisyUIThemes = []string{
	"light", "dark", "cupcake", "bumblebee", "emerald", "corporate",
	"synthwave", "retro", "cyberpunk", "valentine", "halloween",
	"garden", "forest", "aqua", "lofi", "pastel", "fantasy",
	"wireframe", "black", "luxury", "dracula", "cmyk", "autumn",
	"business", "acid", "lemonade", "night", "coffee", "winter",
	"dim", "nord", "sunset",
}

// CheckInstallation checks whether the current directory is a valid PureCore installation.
// Returns (isValid, composeFile, envFile).
func CheckInstallation() (bool, string, string) {
	composeFile := "docker-compose.yml"
	envFile := ".env"
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return false, composeFile, envFile
	}
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return false, composeFile, envFile
	}
	return true, composeFile, envFile
}

// FindComposeCmd returns the docker compose command available on the system.
func FindComposeCmd() string {
	if _, err := exec.LookPath("docker"); err != nil {
		return ""
	}
	// Try docker compose (v2 plugin)
	cmd := exec.Command("docker", "compose", "version")
	if cmd.Run() == nil {
		return "docker compose"
	}
	// Try docker-compose (v1 standalone)
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return "docker-compose"
	}
	return ""
}

// GetRunningContainers returns the number of running containers for the compose file.
func GetRunningContainers(composeFile string) int {
	cmd := FindComposeCmd()
	if cmd == "" {
		return 0
	}
	parts := strings.Fields(cmd)
	args := append(parts, "-f", composeFile, "ps", "-q")
	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count
}

// GetEnv reads a value from the .env file.
func GetEnv(key, defaultValue string) string {
	f, err := os.Open(".env")
	if err != nil {
		return defaultValue
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			return strings.TrimSpace(parts[1])
		}
	}
	return defaultValue
}

// SetEnv updates or adds a key=value pair in the .env file.
func SetEnv(key, value string) error {
	input, err := os.ReadFile(".env")
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	found := false
	newKey := key + "="
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, newKey) {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}
	output := strings.Join(lines, "\n")
	return os.WriteFile(".env", []byte(output), 0o644)
}

// RunCompose runs a docker compose command and returns output.
func RunCompose(composeFile string, args ...string) (string, error) {
	cmd := FindComposeCmd()
	if cmd == "" {
		return "", fmt.Errorf(T("compose_not_found"))
	}
	parts := strings.Fields(cmd)
	allArgs := append(parts, "-f", composeFile)
	allArgs = append(allArgs, args...)
	c := exec.Command(allArgs[0], allArgs[1:]...)
	c.Stderr = nil
	out, err := c.CombinedOutput()
	return string(out), err
}

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// FetchVersions fetches the list of available versions from GitHub releases.
func FetchVersions() ([]string, error) {
	url := "https://api.github.com/repos/zhuchunshu/PureCore/releases?per_page=30"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}
	versions := make([]string, 0, len(releases)+1)
	versions = append(versions, "latest")
	for _, r := range releases {
		v := strings.TrimPrefix(r.TagName, "v")
		if v != "" {
			versions = append(versions, v)
		}
	}
	return versions, nil
}

// RestartServices restarts the docker compose services.
func RestartServices(composeFile string) error {
	_, err := RunCompose(composeFile, "up", "-d", "--remove-orphans")
	return err
}

// PullAndRestart pulls the images for the given version and restarts services.
func PullAndRestart(composeFile, version string) error {
	// Set the version env var for this process
	absDir, _ := filepath.Abs(".")
	os.Setenv("PURECORE_VERSION", version)

	cmd := FindComposeCmd()
	if cmd == "" {
		return fmt.Errorf(T("compose_not_found"))
	}
	parts := strings.Fields(cmd)

	// Pull images
	pullCmd := exec.Command(parts[0], append(parts[1:], "-f", composeFile, "pull")...)
	pullCmd.Dir = absDir
	pullCmd.Env = append(os.Environ(), "PURECORE_VERSION="+version)
	if out, err := pullCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull failed: %v\n%s", err, string(out))
	}

	// Restart services
	upCmd := exec.Command(parts[0], append(parts[1:], "-f", composeFile, "up", "-d", "--remove-orphans")...)
	upCmd.Dir = absDir
	upCmd.Env = append(os.Environ(), "PURECORE_VERSION="+version)
	if out, err := upCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("up failed: %v\n%s", err, string(out))
	}
	return nil
}

// UpdatePrefix updates the admin route prefix in .env.
func UpdatePrefix(newPrefix string) error {
	if err := SetEnv("ADMIN_ROUTE_PREFIX", newPrefix); err != nil {
		return err
	}
	return SetEnv("VITE_ADMIN_ROUTE_PREFIX", newPrefix)
}

// UpdateTheme updates the theme in .env.
func UpdateTheme(newTheme string) error {
	if err := SetEnv("THEME", newTheme); err != nil {
		return err
	}
	return SetEnv("VITE_THEME", newTheme)
}
