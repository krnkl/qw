package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

var semverRegex = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$`)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "verify":
		runVerify()
	case "compare":
		runCompare(os.Args[2:])
	case "bump-type":
		runBumpType(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/semver_tool.go <verify|compare|bump-type> [args]")
}

func runVerify() {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read VERSION file: %v\n", err)
		os.Exit(1)
	}
	version := strings.TrimSpace(string(data))
	if !semverRegex.MatchString(version) {
		fmt.Fprintf(os.Stderr, "Error: VERSION %q does not match semantic versioning format (e.g. 0.1.0)\n", version)
		os.Exit(1)
	}

	newV := "v" + strings.TrimPrefix(version, "v")
	if !semver.IsValid(newV) {
		fmt.Fprintf(os.Stderr, "Error: %q is not a valid semver string\n", newV)
		os.Exit(1)
	}
	fmt.Printf("✓ VERSION format is valid: %s\n", newV)

	out, err := exec.Command("git", "tag", "-l", "v*").Output()
	if err != nil {
		return
	}

	rawTags := strings.Split(strings.TrimSpace(string(out)), "\n")
	var highestTag string
	for _, t := range rawTags {
		t = strings.TrimSpace(t)
		if t == "" || !semver.IsValid(t) {
			continue
		}
		if highestTag == "" || semver.Compare(t, highestTag) > 0 {
			highestTag = t
		}
	}

	if highestTag != "" {
		fmt.Printf("ℹ Highest existing git tag found: %s\n", highestTag)
		cmp := semver.Compare(newV, highestTag)
		if cmp < 0 {
			fmt.Fprintf(os.Stderr, "Error: proposed version %s is lower than existing tag %s (downgrade not allowed)\n", newV, highestTag)
			os.Exit(1)
		} else if cmp == 0 {
			fmt.Printf("ℹ Version %s matches highest existing tag (no new release pending)\n", newV)
		} else {
			fmt.Printf("✓ Version %s is strictly higher than latest tag %s (valid release bump)\n", newV, highestTag)
		}
	} else {
		fmt.Printf("ℹ No previous release tags found. First release will be %s\n", newV)
	}
}

func runCompare(args []string) {
	if len(args) < 2 {
		fmt.Println("0")
		return
	}
	v1 := "v" + strings.TrimPrefix(args[0], "v")
	v2 := "v" + strings.TrimPrefix(args[1], "v")
	fmt.Println(semver.Compare(v1, v2))
}

func parseParts(v string) (int, int, int) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	var maj, min, pat int
	if len(parts) > 0 {
		fmt.Sscanf(parts[0], "%d", &maj)
	}
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &min)
	}
	if len(parts) > 2 {
		fmt.Sscanf(parts[2], "%d", &pat)
	}
	return maj, min, pat
}

func runBumpType(args []string) {
	if len(args) < 2 {
		fmt.Println("unknown")
		return
	}
	oldV := "v" + strings.TrimPrefix(args[0], "v")
	newV := "v" + strings.TrimPrefix(args[1], "v")

	if !semver.IsValid(oldV) || !semver.IsValid(newV) {
		fmt.Println("unknown")
		return
	}

	oldMaj, oldMin, _ := parseParts(oldV)
	newMaj, newMin, _ := parseParts(newV)

	if newMaj > oldMaj {
		fmt.Println("major")
		return
	}
	if newMin > oldMin {
		fmt.Println("minor")
		return
	}
	fmt.Println("patch")
}
