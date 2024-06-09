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
	cmd := exec.Command("bash", "-c", defaultNvmEnv+"&&nvm --version")
	out, err := cmd.Output()
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
				// 添加配置内容到文件末尾
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
	cmd := exec.Command("bash", "-c", defaultNvmEnv+"&& nvm ls-remote |cat -v ")
	out, err := cmd.Output()
	var candidates []Candidate

	if err != nil {
		fmt.Println("Error running command:", err)
		return nil
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		regexPattern := `\b(v[0-9]+\.[0-9]+\.[0-9]+|iojs-v[0-9]+\.[0-9]+\.[0-9]+)\b`
		re := regexp.MustCompile(regexPattern)
		var candidate Candidate
		if strings.Contains(line, "0;34m") {
			candidate.Install = true
		}
		if strings.Contains(line, `0;32m->`) {
			candidate.Use = true
			candidate.Install = true
		}
		matches := re.FindAllString(line, -1)
		if len(matches) == 0 {
			continue
		}
		for _, match := range matches {
			candidate.Identifier = match
		}

		candidates = append(candidates, candidate)
	}
	return candidates
}

func OpenNodeFolder(version string) error {
	cmd := exec.Command("bash", "-c", defaultNvmEnv+"&& nvm which "+version)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running command:", err)
		return nil
	}
	out = []byte(strings.ReplaceAll(string(out), "/bin/node", ""))
	openFolder(strings.TrimSpace(string(out)))
	return nil
}

func InstallNode(version string) {
	cmd := exec.Command("bash", "-c", defaultNvmEnv+"&& nvm install "+version+" && nvm alias node default "+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
	}
}

func UninstallNode(version string) {
	cmd := exec.Command("bash", "-c", defaultNvmEnv+"&& nvm uninstall "+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
	}
}
