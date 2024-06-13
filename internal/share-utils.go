package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Candidate struct {
	Use        bool
	Install    bool
	Identifier string
	Custom     bool
}

func parseVersion(identifier string) (int, int, int, error) {
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(identifier)
	if matches == nil || len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid version format")
	}
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, err
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, err
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, err
	}
	return major, minor, patch, nil
}

func SortCandidates(candidates []Candidate) []Candidate {
	versionRegex := regexp.MustCompile(`\d+\.\d+\.\d+`)

	sort.Slice(candidates, func(i, j int) bool {
		iIsVersion := versionRegex.MatchString(candidates[i].Identifier)
		jIsVersion := versionRegex.MatchString(candidates[j].Identifier)

		if iIsVersion && jIsVersion {
			iMajor, iMinor, iPatch, _ := parseVersion(candidates[i].Identifier)
			jMajor, jMinor, jPatch, _ := parseVersion(candidates[j].Identifier)

			if iMajor != jMajor {
				return iMajor > jMajor
			}
			if iMinor != jMinor {
				return iMinor > jMinor
			}
			return iPatch > jPatch
		}

		if iIsVersion {
			return true
		}
		if jIsVersion {
			return false
		}

		return candidates[i].Identifier > candidates[j].Identifier
	})

	return candidates
}

func CommandExec(commands []string) (string, error) {
	cmd := exec.Command("bash", "-c", strings.Join(commands, " "))
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("error " + strings.Join(commands, " "))
		fmt.Println("Error running command:", err)
		return "", err
	}
	return string(out), nil
}

func containsEnv(filePath string, env string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filePath, err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), env) {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", filePath, err)
	}
	return false
}

func EnvWrite(defaultEnvScript string, provider string, env string) {
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

			if !containsEnv(configFilePath, env) {
				if _, err := file.WriteString("\n" + defaultEnvScript + "\n"); err != nil {
					fmt.Printf("Error writing to %s: %v\n", configFilePath, err)
				} else {
					fmt.Printf("Updated %s\n", configFilePath)
				}
			} else {
				fmt.Printf("%s already contains"+provider+" settings, skipping...\n", configFilePath)
			}
		} else {
			fmt.Printf("%s does not exist, skipping...\n", configFilePath)
		}
	}
}
