package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func get_executable_binaries_path() (string, error) {
	cmd := exec.Command("go", "env", "GOBIN")
	gobin_bytes, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	gobin := strings.TrimSpace(string(gobin_bytes))

	if len(gobin) > 0 {
		return gobin, nil
	}

	cmd = exec.Command("go", "env", "GOPATH")
	gopath_bytes, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	gopath := strings.TrimSpace(string(gopath_bytes))
	if len(gopath) == 0 {
		return "", errors.New("GOPATH and GOPATH are not defined in 'go env' command")
	}

	gobin = fmt.Sprintf("%s/%s", gopath, "bin")

	return gobin, nil
}

func main() {
	gobin, err := get_executable_binaries_path()
	if err != nil {
		fmt.Println("Error while trying to determine the executable binaries path", err)
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
			cmd = exec.Command("go", "install", fmt.Sprintf("%s@latest", binaryPath))
			stdout, err = cmd.CombinedOutput()

			if err != nil {
				fmt.Println("Error when updating package", binaryPath, ", path", binaryPath)
				os.Exit(6)
			}
			fmt.Println("Upgraded package", binaryName)
			fmt.Println(string(stdout))
		}
		if !matched {
			fmt.Println("Output for package", binaryName, "does not contain a \"path\" section")
			os.Exit(5)
		}
	}

	fmt.Println("All packages updated successfully")
}
