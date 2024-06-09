package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Candidate struct {
	Use        bool
	Install    bool
	Identifier string
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
