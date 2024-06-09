package internal

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func JavaVersionList(scriptPath string) []Candidate {
	var javaVersions []Candidate
	out, err := CommandExec([]string{"source " + scriptPath + " && sdk list java"})
	if err != nil {
		fmt.Println("Error running command:", err)
		return javaVersions
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {

		if strings.Contains(line, "====") || strings.Contains(line, "----") || strings.Contains(line, "Vendor") {
			continue
		}

		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			versionInfo := Candidate{
				Use:        strings.TrimSpace(parts[1]) != "",
				Install:    strings.TrimSpace(parts[4]) != "",
				Identifier: strings.TrimSpace(parts[5]),
			}
			javaVersions = append(javaVersions, versionInfo)
		}
	}
	return javaVersions
}

func OtherVersionList(candidate string, scriptPath string) []Candidate {

	out, _ := CommandExec([]string{"source " + scriptPath + " && sdk list " + candidate})
	lines := strings.Split(out, "\n")
	re := regexp.MustCompile(`([>*\s]*)\s*(\d+\.\d+(\.\d+)?(-beta-\d+)?(_\d+)?(-\w+)?(-\w+)?)`)

	var versionInfos []Candidate

	for _, line := range lines {
		if strings.Contains(line, "====") || strings.Contains(line, "----") || strings.Contains(line, "Vendor") {
			continue
		}
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			status := strings.TrimSpace(match[1])
			version := match[2]
			installed := false
			used := false
			if strings.Contains(status, "*") {
				installed = true
			}
			if strings.Contains(status, ">") {
				used = true
			}
			versionInfos = append(versionInfos, Candidate{
				Identifier: version,
				Install:    installed,
				Use:        used,
			})
		}
	}

	return versionInfos
}

func OpenCandidateFolder(candidate string, version, scriptPath string) {
	out, _ := CommandExec([]string{"source " + scriptPath + " && sdk home " + candidate + " " + version})
	openFolder(strings.TrimSpace(out))
}

func CandidateList(scriptPath string) []string {
	out, _ := CommandExec([]string{"source " + scriptPath + " && sdk list"})
	lines := strings.Split(out, "\n")
	re := regexp.MustCompile(`\$ sdk install (\S+)`)

	// Slice to hold the install commands
	var installCommands []string

	// Iterate over the lines and extract the installation commands
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			installCommands = append(installCommands, matches[1])
		}
	}
	return installCommands
}

func UseCandidate(candidate string, version string, scriptPath string) {
	fmt.Println("Installing", candidate, version)
	_, err := CommandExec([]string{"source " + scriptPath + " && sdk install " + candidate + " " + version + " && sdk default " + candidate + " " + version})
	if err == nil {
		fmt.Println("Installed", candidate, version)
	}
}

func UninstallCandidate(candidate string, version string, scriptPath string) {
	fmt.Println("UnInstalling", candidate, version)
	_, err := CommandExec([]string{"source " + scriptPath + " && sdk uninstall " + candidate + " " + version})
	if err == nil {
		fmt.Println("UnInstalled", candidate, version)
	}
}

func openFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func InstallSDKMan() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return fmt.Errorf("HOME environment variable is not set")
	}

	sdkManPath := filepath.Join(homeDir, ".sdkman")
	if FileExists(sdkManPath) {
		fmt.Println("SDKMan already installed")
		return nil
	}
	beeep.Notify("SDKMan Installation", "SDKMan is not installed, Installing SDKMan", "")
	output, err := CommandExec([]string{"curl -s \"https://get.sdkman.io\" | bash"})
	if err != nil {
		fmt.Printf("Error running command: %v\nOutput: %s\n", err, string(output))
		return err
	}
	fmt.Println("SDKMan installed successfully")
	beeep.Notify("SDKMan Installation", "SDKMan installed successfully", "")
	return nil
}

func SDKManVersion(scriptPath string) string {
	cmd := exec.Command("bash", "-c", "source "+scriptPath+" && sdk version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\nOutput: %s\n", err, string(output))
		return ""
	}

	return strings.TrimSpace(string(output))
}

func SDKManUpdate(scriptPath string) error {
	cmd := exec.Command("bash", "-c", "source "+scriptPath+" && sdk update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\nOutput: %s\n", err, string(output))
		return err
	}

	return nil
}
