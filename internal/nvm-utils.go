package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	defaultNvmEnv = `export NVM_DIR="$HOME/.nvm"; [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"; [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"`
)

func InstallNVM() {
	// Install NVM
	NVMEnvWrite()
	out, err := CommandExec([]string{defaultNvmEnv + "&& nvm --version"})
	if err != nil {
		fmt.Println("Error running command:", err)
		fmt.Println("Installing NVM")

		exec.Command("bash", "-c", "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash").Run()

		return
	}
	fmt.Println("NVM is already installed", string(out))
}

func NVMEnvWrite() {
	shellConfigFiles := []string{".bashrc", ".zshrc", ".profile"}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		return
	}

	for _, configFile := range shellConfigFiles {
		configFilePath := filepath.Join(homeDir, configFile)
		if _, err := os.Stat(configFilePath); err == nil {
			file, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error opening %s: %v\n", configFilePath, err)
				continue
			}
			defer file.Close()

			if !containsNvmEnv(configFilePath) {

				if _, err := file.WriteString("\n" + defaultNvmEnv + "\n"); err != nil {
					fmt.Printf("Error writing to %s: %v\n", configFilePath, err)
				} else {
					fmt.Printf("Updated %s\n", configFilePath)
				}
			} else {
				fmt.Printf("%s already contains NVM settings, skipping...\n", configFilePath)
			}
		} else {
			fmt.Printf("%s does not exist, skipping...\n", configFilePath)
		}
	}
}

// 检查配置文件是否包含 defaultNvmEnv
func containsNvmEnv(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filePath, err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "export NVM_DIR") {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", filePath, err)
	}
	return false
}

func NodeVersionList() []Candidate {
	// List Node versions
	var candidates []Candidate
	out, err := CommandExec([]string{defaultNvmEnv + "&& nvm ls-remote"})
	if err != nil {
		return candidates
	}
	lines := strings.Split(out, "\n")
	installedMap := NodeLocalInstallList()
	for _, line := range lines {
		regexPattern := `\b(v[0-9]+\.[0-9]+\.[0-9]+)\b`
		re := regexp.MustCompile(regexPattern)
		var candidate Candidate
		matches := re.FindAllString(line, -1)
		if len(matches) == 0 {
			continue
		}
		for _, match := range matches {
			if _, ok := installedMap[match]; ok {
				candidate = installedMap[match]
			} else {
				candidate.Identifier = match
				candidate.Install = false
			}
		}

		candidates = append(candidates, candidate)
	}
	return candidates
}

func OpenNodeFolder(version string) {
	out, _ := CommandExec([]string{defaultNvmEnv + "&& nvm which " + version})
	out = strings.ReplaceAll(out, "/bin/node", "")
	openFolder(strings.TrimSpace(out))
}

func InstallNode(version string) {
	fmt.Println("Installing Node version", version)
	_, _ = CommandExec([]string{defaultNvmEnv + "&& nvm install " + version + " && nvm alias default " + version})
	fmt.Println("Installed Node version", version)
}

func UninstallNode(version string) {
	fmt.Println("Uninstalling Node version", version)
	_, _ = CommandExec([]string{defaultNvmEnv + "&& nvm uninstall " + version})
	fmt.Println("Uninstalled Node version", version)
}

func NVMVersion() string {
	out, err := CommandExec([]string{defaultNvmEnv + "&& nvm --version"})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func NodeLocalInstallList() map[string]Candidate {
	var installCandidates = make(map[string]Candidate)
	out, _ := CommandExec([]string{defaultNvmEnv + "&& nvm ls node"})
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		var candidate Candidate

		regexPattern := `\b(v[0-9]+\.[0-9]+\.[0-9]+)\b`
		re := regexp.MustCompile(regexPattern)
		matches := re.FindAllString(line, -1)
		if len(matches) == 0 {
			continue
		}
		for _, match := range matches {
			candidate.Identifier = match
			candidate.Install = true
		}
		if strings.Contains(line, `->`) {
			candidate.Use = true
		}
		if candidate.Identifier != "" {
			installCandidates[candidate.Identifier] = candidate
		}
	}
	return installCandidates
}
