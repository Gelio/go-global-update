package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	gobin := os.Getenv("GOBIN")
	if len(gobin) == 0 {
		fmt.Println("GOBIN environment variable is not defined")
		os.Exit(1)
	}

	if err := os.Chdir(gobin); err != nil {
		fmt.Println("Error when changing directory to GOBIN", err)
		os.Exit(2)
	}

	fileinfos, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error when reading the GOBIN directory", err)
		os.Exit(3)
	}

	for _, fileinfo := range fileinfos {
		binaryName := fileinfo.Name()
		fmt.Println("Upgrading package", binaryName)
		cmd := exec.Command("go", "version", "-m", binaryName)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println("Error when reading version information for package", binaryName, err)
			os.Exit(4)
		}
		stdoutString := string(stdout)

		lines := strings.Split(stdoutString, "\n")
		matched := false

		for _, l := range lines {
			r := regexp.MustCompile(`\bpath\s+(.*)$`)
			matches := r.FindStringSubmatch(l)
			if len(matches) == 0 {
				continue
			}
			matched = true

			binaryPath := matches[1]
			cmd = exec.Command("go", "get", "-u", binaryPath)
			stdout, err = cmd.Output()

			if err != nil {
				fmt.Println("Error when updating package", binaryPath, ", path", binaryPath)
				os.Exit(6)
			}
			fmt.Println("Upgraded package", binaryName)
			fmt.Print(stdout)
		}
		if !matched {
			fmt.Println("Output for package", binaryName, "does not contain a \"path\" section")
			os.Exit(5)
		}
	}

	fmt.Println("All packages updated successfully")
}
