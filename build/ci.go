// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// +build none

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/aidan-kwon/toy-klaytn/utils/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (

	// Files that end up in the klay*.zip archive.
	klayArchiveFiles = []string{
		"COPYING",
		executablePath("klay"),
	}

	// Files that end up in the klay-alltools*.zip archive.
	allToolsArchiveFiles = []string{
		"COPYING",
		executablePath("klay"),
	}

	// A debian package is created for all executables listed here.
	debExecutables = []debExecutable{
		{
			Name:        "klay",
			Description: "Klaytn CLI client.",
		},
	}

	// Distros for which packages are created.
	// Note: vivid is unsupported because there is no golang-1.6 package for it.
	// Note: wily is unsupported because it was officially deprecated on lanchpad.
	// Note: yakkety is unsupported because it was officially deprecated on lanchpad.
	// Note: zesty is unsupported because it was officially deprecated on lanchpad.
	debDistros = []string{"trusty", "xenial", "artful", "bionic"}
)

var GOBIN, _ = filepath.Abs(filepath.Join("build", "bin"))

func executablePath(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(GOBIN, name)
}

func main() {
	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("build", "ci.go")); os.IsNotExist(err) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "install":
		doInstall(os.Args[2:])
	case "test":
		doTest(os.Args[2:])
	case "fmt":
		doFmt(os.Args[2:])
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

// Compiling

func doInstall(cmdline []string) {
	var (
		arch = flag.String("arch", "", "Architecture to cross build for")
		cc   = flag.String("cc", "", "C compiler to cross build with")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	// Check Go version. People regularly open issues about compilation
	// failure with outdated Go. This should save them the trouble.
	if !strings.Contains(runtime.Version(), "devel") {
		// Figure out the minor version number since we can't textually compare (1.10 < 1.9)
		var minor int
		fmt.Sscanf(strings.TrimPrefix(runtime.Version(), "go1."), "%d", &minor)

		if minor < 9 {
			log.Println("You have Go version", runtime.Version())
			log.Println("klaytn requires at least Go version 1.9 and cannot")
			log.Println("be compiled with an earlier version. Please upgrade your Go installation.")
			os.Exit(1)
		}
	}
	// Compile packages given as arguments, or everything if there are no arguments.
	packages := []string{"./..."}
	if flag.NArg() > 0 {
		packages = flag.Args()
	}

	if *arch == "" || *arch == runtime.GOARCH {
		goinstall := goTool("install", buildFlags(env)...)
		goinstall.Args = append(goinstall.Args, "-v")
		//goinstall.Args = append(goinstall.Args, "-race")
		goinstall.Args = append(goinstall.Args, packages...)
		build.MustRun(goinstall)
		return
	}
	// If we are cross compiling to ARMv5 ARMv6 or ARMv7, clean any previous builds
	if *arch == "arm" {
		os.RemoveAll(filepath.Join(runtime.GOROOT(), "pkg", runtime.GOOS+"_arm"))
		for _, path := range filepath.SplitList(build.GOPATH()) {
			os.RemoveAll(filepath.Join(path, "pkg", runtime.GOOS+"_arm"))
		}
	}
	// Seems we are cross compiling, work around forbidden GOBIN
	goinstall := goToolArch(*arch, *cc, "install", buildFlags(env)...)
	goinstall.Args = append(goinstall.Args, "-v")
	goinstall.Args = append(goinstall.Args, []string{"-buildmode", "archive"}...)
	goinstall.Args = append(goinstall.Args, packages...)
	build.MustRun(goinstall)

	if cmds, err := ioutil.ReadDir("cmd"); err == nil {
		for _, cmd := range cmds {
			pkgs, err := parser.ParseDir(token.NewFileSet(), filepath.Join(".", "cmd", cmd.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				log.Fatal(err)
			}
			for name := range pkgs {
				if name == "main" {
					gobuild := goToolArch(*arch, *cc, "build", buildFlags(env)...)
					gobuild.Args = append(gobuild.Args, "-v")
					gobuild.Args = append(gobuild.Args, []string{"-o", executablePath(cmd.Name())}...)
					gobuild.Args = append(gobuild.Args, "."+string(filepath.Separator)+filepath.Join("cmd", cmd.Name()))
					build.MustRun(gobuild)
					break
				}
			}
		}
	}
}

func buildFlags(env build.Environment) (flags []string) {
	var ld []string
	if env.Commit != "" {
		ld = append(ld, "-X", "main.gitCommit="+env.Commit)
		ld = append(ld, "-X", "github.com/aidan-kwon/toy-klaytn/cmd/utils/nodecmd.gitCommit="+env.Commit)
	}
	if env.Tag != "" {
		ld = append(ld, "-X", "github.com/aidan-kwon/toy-klaytn/cmd/utils/nodecmd.gitTag="+env.Tag)
	}
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}

	if env.IsDisabledSymbolTable {
		ld = append(ld, "-s")
	}
	if env.IsStaticLink {
		// Pass the static link flag to the external linker.
		// By default, cmd/link will use external linking mode when non-standard cgo packages are involved.
		ld = append(ld, "-linkmode", "external", "-extldflags", "-static")
	}
	if env.IsKlaytnRaceDetectionOn {
		flags = append(flags, "-race")
	}
	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	return flags
}

func goTool(subcmd string, args ...string) *exec.Cmd {
	return goToolArch(runtime.GOARCH, os.Getenv("CC"), subcmd, args...)
}

func goToolArch(arch string, cc string, subcmd string, args ...string) *exec.Cmd {
	cmd := build.GoTool(subcmd, args...)
	cmd.Env = []string{"GOPATH=" + build.GOPATH()}
	if arch == "" || arch == runtime.GOARCH {
		cmd.Env = append(cmd.Env, "GOBIN="+GOBIN)
	} else {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		cmd.Env = append(cmd.Env, "GOARCH="+arch)
	}
	if cc != "" {
		cmd.Env = append(cmd.Env, "CC="+cc)
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") || strings.HasPrefix(e, "GOBIN=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	return cmd
}

// Running The Tests
//
// "tests" also includes static analysis tools such as vet.

func doTest(cmdline []string) {
	var (
		parallel = flag.Int("p", 0, "The number of parallel test executions (default: the number of CPUs available)")
		excludes = flag.String("exclude", "", "Comma-separated top-level directories to be excluded in test")
	)
	flag.CommandLine.Parse(cmdline)
	env := build.Env()

	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	if *excludes != "" {
		packages = build.ExcludePackages(packages, strings.Split(*excludes, ","))
	}

	// Run analysis tools before the tests.
	build.MustRun(goTool("vet", packages...))

	// Run the actual tests.
	gotest := goTool("test", buildFlags(env)...)
	if *parallel != 0 {
		gotest.Args = append(gotest.Args, "-p", strconv.Itoa(*parallel))
	}
	gotest.Args = append(gotest.Args, "--timeout=30m")
	gotest.Args = append(gotest.Args, packages...)
	build.MustRun(gotest)
}

func doFmt(cmdline []string) {
	// runs gometalinter on requested packages
	flag.CommandLine.Parse(cmdline)

	packages := []string{"./..."}
	if len(flag.CommandLine.Args()) > 0 {
		packages = flag.CommandLine.Args()
	}

	lintBin := installLinter()

	// Run fast linters batched together
	configs := []string{
		"run",
		"--tests",
		"--disable-all",
		"--enable=gofmt",
		"--timeout=2m",
	}
	build.MustRunCommand(lintBin, append(configs, packages...)...)
}

func archiveBasename(arch string, env build.Environment) string {
	platform := runtime.GOOS + "-" + arch
	if arch == "arm" {
		platform += os.Getenv("GOARM")
	}
	if arch == "android" {
		platform = "android-all"
	}
	if arch == "ios" {
		platform = "ios-all"
	}
	return platform + "-" + archiveVersion(env)
}

func archiveVersion(env build.Environment) string {
	version := build.VERSION()
	if isUnstableBuild(env) {
		version += "-unstable"
	}
	if env.Commit != "" {
		version += "-" + env.Commit[:8]
	}
	return version
}

func archiveUpload(archive string, blobstore string, signer string) error {
	// If signing was requested, generate the signature files
	if signer != "" {
		pgpkey, err := base64.StdEncoding.DecodeString(os.Getenv(signer))
		if err != nil {
			return fmt.Errorf("invalid base64 %s", signer)
		}
		if err := build.PGPSignFile(archive, archive+".asc", string(pgpkey)); err != nil {
			return err
		}
	}
	return nil
}

// skips archiving for some build configurations.
func maybeSkipArchive(env build.Environment) {
	if env.IsPullRequest {
		log.Printf("skipping because this is a PR build")
		os.Exit(0)
	}
	if env.IsCronJob {
		log.Printf("skipping because this is a cron job")
		os.Exit(0)
	}
	if env.Branch != "master" && !strings.HasPrefix(env.Tag, "v1.") {
		log.Printf("skipping because branch %q, tag %q is not on the whitelist", env.Branch, env.Tag)
		os.Exit(0)
	}
}

func makeWorkdir(wdflag string) string {
	var err error
	if wdflag != "" {
		err = os.MkdirAll(wdflag, 0744)
	} else {
		wdflag, err = ioutil.TempDir("", "klay-build-")
	}
	if err != nil {
		log.Fatal(err)
	}
	return wdflag
}

func isUnstableBuild(env build.Environment) bool {
	if env.Tag != "" {
		return false
	}
	return true
}

func installLinter() string {
	lintBin := filepath.Join(build.GOPATH(), "bin", "golangci-lint")

	_, err := exec.LookPath(lintBin)
	if err != nil {
		fmt.Println("Installing golangci-lint.")

		cmdCurl := exec.Command("curl", "-sSfL", "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh")
		cmdSh := exec.Command("sh", "-s", "--", "-b", filepath.Join(build.GOPATH(), "bin"), "v1.24.0")
		cmdSh.Stdin, err = cmdCurl.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(">>>", strings.Join(cmdCurl.Args, " "))
		if err := cmdCurl.Start(); err != nil {
			log.Fatal(err)
		}

		build.MustRun(cmdSh)

		if err := cmdCurl.Wait(); err != nil {
			log.Fatal(err)
		}
	}
	return lintBin
}


type debMetadata struct {
	Env build.Environment

	// klaytn version being built. Note that this
	// is not the debian package version. The package version
	// is constructed by VersionString.
	Version string

	Author       string // "name <email>", also selects signing key
	Distro, Time string
	Executables  []debExecutable
}

type debExecutable struct {
	Name, Description string
}

func newDebMetadata(distro, author string, env build.Environment, t time.Time) debMetadata {
	if author == "" {
		// No signing key, use default author.
		author = "Klaytn Builds <infra@groundx.xyz>"
	}
	return debMetadata{
		Env:         env,
		Author:      author,
		Distro:      distro,
		Version:     build.VERSION(),
		Time:        t.Format(time.RFC1123Z),
		Executables: debExecutables,
	}
}

// Name returns the name of the metapackage that depends
// on all executable packages.
func (meta debMetadata) Name() string {
	if isUnstableBuild(meta.Env) {
		return "klaytn-unstable"
	}
	return "klaytn"
}

// VersionString returns the debian version of the packages.
func (meta debMetadata) VersionString() string {
	vsn := meta.Version
	if meta.Env.Buildnum != "" {
		vsn += "+build" + meta.Env.Buildnum
	}
	if meta.Distro != "" {
		vsn += "+" + meta.Distro
	}
	return vsn
}

// ExeList returns the list of all executable packages.
func (meta debMetadata) ExeList() string {
	names := make([]string, len(meta.Executables))
	for i, e := range meta.Executables {
		names[i] = meta.ExeName(e)
	}
	return strings.Join(names, ", ")
}

// ExeName returns the package name of an executable package.
func (meta debMetadata) ExeName(exe debExecutable) string {
	if isUnstableBuild(meta.Env) {
		return exe.Name + "-unstable"
	}
	return exe.Name
}

// ExeConflicts returns the content of the Conflicts field
// for executable packages.
func (meta debMetadata) ExeConflicts(exe debExecutable) string {
	if isUnstableBuild(meta.Env) {
		// Set up the conflicts list so that the *-unstable packages
		// cannot be installed alongside the regular version.
		//
		// https://www.debian.org/doc/debian-policy/ch-relationships.html
		// is very explicit about Conflicts: and says that Breaks: should
		// be preferred and the conflicting files should be handled via
		// alternates. We might do this eventually but using a conflict is
		// easier now.
		return "klaytn, " + exe.Name
	}
	return ""
}
