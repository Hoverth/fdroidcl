// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const cmdName = "fdroidcl"

const version = "v0.7.0"

func subdir(dir, name string) string {
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create dir '%s': %v\n", p, err)
	}
	return p
}

func mustCache() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("TODO: return an error")
	}
	return subdir(dir, cmdName)
}

func mustData() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("TODO: return an error")
	}
	return subdir(dir, cmdName)
}

func configPath() string {
	return filepath.Join(mustData(), "config.toml")
}

type repo struct {
	ID      string `toml:"id"`
	URL     string `toml:"url"`
	Enabled bool   `toml:"enabled"`
}

type setup struct {
	ID    string   `toml:"id"`
	Apps  []string `toml:"apps"`
	Repos []string `toml:"repos"`
}

type userConfig struct {
	Repos  []repo  `toml:"repos"`
	Setups []setup `toml:"setups"`
}

var config = userConfig{
	Repos: []repo{
		{
			ID:      "f-droid",
			URL:     "https://f-droid.org/repo",
			Enabled: true,
		},
		{
			ID:      "f-droid-archive",
			URL:     "https://f-droid.org/archive",
			Enabled: false,
		},
	},
	Setups: []setup{},
}

func readConfig() error {
	f, err := os.Open(configPath())
	if err != nil {
		// ignore error, if file does not exist
		return nil
	}
	defer f.Close()
	fileConfig := userConfig{}
	err = toml.NewDecoder(f).Decode(&fileConfig)
	if err != nil {
		return err
	}
	config = fileConfig
	return nil
}

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args []string) error

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short, single-line description.
	Short string

	// Long is an optional longer version of the Short description.
	Long string

	Fset flag.FlagSet
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) usage() {
	fmt.Fprintf(os.Stderr, "usage: %s %s\n\n", cmdName, c.UsageLine)
	if c.Long == "" {
		fmt.Fprintf(os.Stderr, "%s.\n", c.Short)
	} else {
		fmt.Fprint(os.Stderr, c.Long)
	}
	anyFlags := false
	c.Fset.VisitAll(func(f *flag.Flag) { anyFlags = true })
	if anyFlags {
		fmt.Fprintf(os.Stderr, "\nAvailable options:\n")
		c.Fset.PrintDefaults()
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-h] <command> [<args>]\n\n", cmdName)
		fmt.Fprintf(os.Stderr, "Available commands:\n")
		maxUsageLen := 0
		for _, c := range commands {
			if len(c.UsageLine) > maxUsageLen {
				maxUsageLen = len(c.UsageLine)
			}
		}
		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "   %s%s  %s\n", c.UsageLine,
				strings.Repeat(" ", maxUsageLen-len(c.UsageLine)), c.Short)
		}
		fmt.Fprintf(os.Stderr, `
An appid is just an app's unique package name. A specific version of an app can
be selected by following the appid with a colon and the version code. The
'search' and 'show' commands can be used to find these strings. For example:

	$ fdroidcl search redreader
	$ fdroidcl show org.quantumbadger.redreader
	$ fdroidcl install org.quantumbadger.redreader:85
`)
		fmt.Fprintf(os.Stderr, "\nUse %s <command> -h for more information.\n", cmdName)
	}
}

// Commands lists the available commands.
var commands = []*Command{
	cmdUpdate,
	cmdSearch,
	cmdShow,
	cmdInstall,
	cmdUninstall,
	cmdDownload,
	cmdDevices,
	cmdScan,
	cmdList,
	cmdRepo,
	cmdSetup,
	cmdClean,
	cmdDefaults,
	cmdVersion,
}

var cmdVersion = &Command{
	UsageLine: "version",
	Short:     "Print version information",
	Run: func(args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("no arguments allowed")
		}
		fmt.Println(version)
		return nil
	},
}

func main() {
	os.Exit(main1())
}

func main1() int {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		return 2
	}

	cmdName := args[0]
	for _, cmd := range commands {
		if cmd.Name() != cmdName {
			continue
		}
		cmd.Fset.Init(cmdName, flag.ContinueOnError)
		cmd.Fset.Usage = cmd.usage
		if err := cmd.Fset.Parse(args[1:]); err != nil {
			if err != flag.ErrHelp {
				fmt.Fprintf(os.Stderr, "flag: %v\n", err)
				cmd.Fset.Usage()
			}
			return 2
		}

		err := readConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config %s: %v\n", configPath(), err)
			return 1
		}

		if err := cmd.Run(cmd.Fset.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", cmdName, err)
			return 1
		}
		return 0
	}
	fmt.Fprintf(os.Stderr, "Unrecognised command '%s'\n\n", cmdName)
	flag.Usage()
	return 2
}
