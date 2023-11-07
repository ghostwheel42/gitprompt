package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akupila/gitprompt"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	goversion = "unknown"
)

const defaultFormat = "#B([@b#R%h][#y ›%s][#m ↓%b][#m ↑%a][#r x%c][#g +%m][#y %u]#B) "

type formatFlag struct {
	set   bool
	value string
}

func (f *formatFlag) Set(v string) error {
	f.set = true
	f.value = v
	return nil
}

func (f *formatFlag) String() string {
	if f.set {
		return f.value
	}

	if envVar := os.Getenv("GITPROMPT_FORMAT"); envVar != "" {
		return envVar
	}

	return defaultFormat
}

func showHelp() {

	var exampleStatus = &gitprompt.GitStatus{
		Branch:    "master",
		Sha:       "0455b83f923a40f0b485665c44aa068bc25029f5",
		Untracked: 1,
		Modified:  2,
		Staged:    3,
		Conflicts: 4,
		Ahead:     5,
		Behind:    6,
		Stashed:   7,
		Upstream:  "origin/master",
		Clean:     true,
	}

	out := flag.CommandLine.Output()

	fmt.Fprintf(out, "Usage:\n\n")
	flag.CommandLine.PrintDefaults()

	example := gitprompt.Print(exampleStatus, defaultFormat, false)

	fmt.Fprintf(out, `
  Default format: %q
  Example result: %s

  Data:
    %%h  Current branch or first 7 hex-digits of SHA1
    %%H  Current branch or first 7 hex-digits of SHA1 prefixed by :
    %%s  Number of files staged
    %%b  Number of commits behind remote
    %%a  Number of commits ahead of remote
    %%c  Number of conflicts
    %%m  Number of files modified
    %%u  Number of untracked files
    %%S  Number of stashed changes
    %%U  Name of tracked upstream branch

  Enablers force-enable a group:
    %%C  Enable group when clean
    %%D  Enable group when not clean (or dirty)
    %%O  Enable group when outdated
    %%L  Enable group when latest (or up to date)
    %%l  Enable group when there's no upstream (local repository)
    %%e  Enable group when last group was not enabled

  Colors:
    #k  Black
    #r  Red
    #g  Green
    #y  Yellow
    #b  Blue
    #m  Magenta
    #c  Cyan
    #w  White
    #K  Highlight Black
    #R  Highlight Red
    #G  Highlight Green
    #Y  Highlight Yellow
    #B  Highlight Blue
    #M  Highlight Magenta
    #C  Highlight Cyan
    #W  Highlight White
    #_  Reset color
    #>  Leak color

  Text attributes:
    @b  Set bold
    @B  Clear bold
    @f  Set faint/dim color
    @F  Clear faint/dim color
    @i  Set italic
    @I  Clear italic
    @_  Reset attributes
    @>  Leak attributes

`, defaultFormat, example)
}

func main() {

	var format formatFlag

	v := flag.Bool("version", false, "Print version information")
	zsh := flag.Bool("zsh", false, "Print zsh width control characters")
	flag.Usage = showHelp
	flag.Var(&format, "format", "Define output format (see below)")
	flag.Parse()

	if *v {
		fmt.Printf(
			"Version:    %s\nCommit:     %s\nBuild date: %s\nGo version: %s\n",
			version, commit, date, goversion)
		os.Exit(0)
	}

	s, err := gitprompt.Parse()
	if err != nil {
		if !*zsh {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	if s == nil {
		return
	}

	fmt.Print(gitprompt.Print(s, format.String(), *zsh))

}
