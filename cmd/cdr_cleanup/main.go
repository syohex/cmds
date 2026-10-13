package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var dryRun = false
var entryRegexp = regexp.MustCompile(`\$'(.*)'`)

func init() {
	flag.BoolVar(&dryRun, "dry-run", false, "only print deleting entries")
	flag.Parse()
}

func readCdrEntries(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	dirs := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		matches := entryRegexp.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		dirs = append(dirs, matches[1])
	}

	return dirs, nil
}

func compileFilteredPatterns() ([]*regexp.Regexp, error) {
	defaultPattern := `(?:\.|node_modules|_?vendor|tmp|temp)`

	patterns := flag.Args()
	patterns = append(patterns, defaultPattern)

	regexps := []*regexp.Regexp{}
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("cannot compile pattern '%s': %w", pattern, err)
		}

		regexps = append(regexps, re)
	}

	return regexps, nil
}

func filterEntries(dirs []string, regexps []*regexp.Regexp) []string {
	keepEntries := []string{}
	for _, dir := range dirs {
		ok := true
		for _, re := range regexps {
			if re.MatchString(dir) {
				ok = false
				break
			}
		}

		if ok {
			// keep only existing directories
			if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
				keepEntries = append(keepEntries, dir)
			}
		}
	}

	return keepEntries
}

func writeBackEntries(cdrFile string, dirs []string) error {
	f, err := os.OpenFile(cdrFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s to write: %w", cdrFile, err)
	}
	defer f.Close()

	uniqueDirs := map[string]struct{}{}
	for _, dir := range dirs {
		uniqueDirs[dir] = struct{}{}
	}

	for dir := range uniqueDirs {
		_, err := f.WriteString(fmt.Sprintf("$'%s'\n", dir))
		if err != nil {
			return err
		}
	}

	return nil
}

func _main() int {
	regexps, err := compileFilteredPatterns()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get home directory: %v", err)
		return 1
	}

	cdrFile := filepath.Join(home, ".chpwd-recent-dirs")

	dirs, err := readCdrEntries(cdrFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read cdr entries: %v", err)
		return 1
	}

	keeps := filterEntries(dirs, regexps)

	if dryRun {
		for _, keep := range keeps {
			fmt.Println(keep)
		}
		return 0
	}

	if err := writeBackEntries(cdrFile, keeps); err != nil {
		fmt.Fprintf(os.Stderr, "cannot write back entries: %v", err)
		return 1
	}

	deletedEntries := len(dirs) - len(keeps)
	fmt.Printf("Delete %d entries\n", deletedEntries)
	return 0
}

func main() {
	os.Exit(_main())
}
