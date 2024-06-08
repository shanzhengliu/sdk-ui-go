package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

type Candidate struct {
	Use        bool
	Install    bool
	Identifier string
}

func VersionList(candidate string, os string) []string {
	resp, err := http.Get("https://api.sdkman.io/2/candidates/" + candidate + "/" + os + "/versions/all")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(body), ",")
}

func JavaVersionList() []Candidate {
	args := []string{"-c", "source " + sdkmanInitScript + " && sdk list java"}
	cmd := exec.Command("bash", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running command:", err)
		return nil

	}
	lines := strings.Split(string(out), "\n")

	var javaVersions []Candidate

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
	for _, version := range javaVersions {
		fmt.Println(version)

	}
	return javaVersions
}

func OtherVersionList(candidate string) []Candidate {
	args := []string{"-c", "source " + sdkmanInitScript + " && sdk list " + candidate}
	cmd := exec.Command("bash", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running command:", err)
		return nil
	}

	lines := strings.Split(string(out), "\n")

	re := regexp.MustCompile(`([>* ]?)\s*(\d+\.\d+\.\d+(-beta-\d+)?)`)

	var versionInfos []Candidate

	for _, line := range lines {
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

	for _, info := range versionInfos {
		fmt.Println(info)
	}

	return versionInfos
}

func CandidateList() []string {
	args := []string{"-c", "source " + sdkmanInitScript + " && sdk list"}
	cmd := exec.Command("bash", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running command:", err)
		return nil
	}
	lines := strings.Split(string(out), "\n")
	re := regexp.MustCompile(`\$ sdk install (\S+)`)

	// Slice to hold the install commands
	var installCommands []string

	// Iterate over the lines and extract the install commands
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			installCommands = append(installCommands, matches[1])
		}
	}

	fmt.Println(installCommands)
	return nil
}

func UseCandidate(candidate string, version string) error {
	fmt.Println("Installing", candidate, version)
	args := []string{"-c", "source " + sdkmanInitScript + " && sdk install " + candidate + " " + version + " && sdk use " + candidate + " " + version}
	cmd := exec.Command("bash", args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return err
	}
	fmt.Println("Installed", candidate, version)
	return nil
}
