// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

var cmdDefaults = &Command{
	UsageLine: "defaults",
	Short:     "Reset to the default settings",
}

func init() {
	cmdDefaults.Run = runDefaults
}

func runDefaults(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("no arguments allowed")
	}
	return writeConfig(&config)
}

func writeConfig(c *userConfig) error {
	b, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot encode config: %v", err)
	}
	f, err := os.Create(configPath())
	if err != nil {
		return fmt.Errorf("cannot create config file: %v", err)
	}
	_, err = f.Write(b)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
