package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/* This script evists because godep deprecated -copy=false, and I really
 * don't agree that importing the actual source code for migemogrep is the
 * correct choice
 */

var pwd string

func init() {
	var err error
	if pwd, err = os.Getwd(); err != nil {
		panic(err)
	}
}

func main() {
	switch os.Args[1] {
	case "deps":
		setupDeps()
	case "build":
		setupDeps()
		buildBinaries()
	default:
		panic("Unknown action: " + os.Args[1])
	}
}

func setupDeps() {
	deps := map[string]string{
		"github.com/koron/gelatin":  "21a9ebd1a4bf74f14044518e75aeb6e8814b581f",
		"github.com/koron/gomigemo": "46cc93985b8a41aa35737a59a9e700012309e008",
	}

	var err error

	for dir, hash := range deps {
		repo := repoURL(dir)
		dir = filepath.Join("src", dir)
		if _, err = os.Stat(dir); err != nil {
			if err = run("git", "clone", repo, dir); err != nil {
				panic(err)
			}
		}

		if err = os.Chdir(dir); err != nil {
			panic(err)
		}

		if err = run("git", "reset", "--hard"); err != nil {
			panic(err)
		}

		if err = run("git", "checkout", "master"); err != nil {
			panic(err)
		}

		if err = run("git", "pull"); err != nil {
			panic(err)
		}

		if err = run("git", "checkout", hash); err != nil {
			panic(err)
		}

		if err = os.Chdir(pwd); err != nil {
			panic(err)
		}
	}

	linkMigemogrepdir()
}

func linkMigemogrepdir() string {
	var err error

	// Link src/github.com/migemogrep/migemogrep to updir
	migemogrepdir := filepath.Join("src", "github.com", "peco", "migemogrep")
	parent := filepath.Dir(migemogrepdir)
	if _, err = os.Stat(parent); err != nil {
		if err = os.MkdirAll(parent, 0777); err != nil {
			panic(err)
		}
	}

	updir, err := filepath.Abs(filepath.Join(pwd, ".."))
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(migemogrepdir); err != nil {
		if err = os.Symlink(updir, migemogrepdir); err != nil {
			panic(err)
		}
	}

	return migemogrepdir
}

func buildBinaries() {
	var err error

	migemogrepdir := linkMigemogrepdir()
	if err = os.Chdir(migemogrepdir); err != nil {
		panic(err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		gopath = strings.Join([]string{pwd, gopath}, string([]rune{filepath.ListSeparator}))
	}
	os.Setenv("GOPATH", gopath)

	goxcArgs := []string{
		"-tasks", "xc archive",
		"-bc", "linux windows darwin",
		"-d", os.Args[2],
		"-resources-include", "README*,Changes",
	}
	if err = run("goxc", goxcArgs...); err != nil {
		panic(err)
	}
}

func run(name string, args ...string) error {
	splat := []string{name}
	splat = append(splat, args...)
	log.Printf("---> Running %v...\n", splat)
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	for _, line := range strings.Split(string(out), "\n") {
		log.Print(line)
	}
	log.Println("")
	log.Println("<---DONE")
	return err
}

func repoURL(spec string) string {
	return "https://" + spec + ".git"
}