// Copyright (c) 2018, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"mvdan.cc/fdroidcl/adb"
)

// chosenApp is the app that will be installed and uninstalled on a connected
// device. This one was chosen because it's tiny, requires no permissions, and
// should be compatible with every device.
//
// It also stores no data, so it is fine to uninstall it and the user won't lose
// any data.
const chosenApp = "org.vi_server.red_screen"

func TestCommands(t *testing.T) {
	return
	url := config.Repos[0].URL
	client := http.Client{Timeout: 2 * time.Second}
	if _, err := client.Get(url); err != nil {
		t.Skipf("skipping since %s is unreachable: %v", url, err)
	}

	dir, err := ioutil.TempDir("", "fdroidcl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	mustSucceed := func(t *testing.T, wantRe, negRe string, cmd *Command, args ...string) {
		mustRun(t, true, wantRe, negRe, cmd, args...)
	}
	mustFail := func(t *testing.T, wantRe, negRe string, cmd *Command, args ...string) {
		mustRun(t, false, wantRe, negRe, cmd, args...)
	}

	if err := startAdbIfNeeded(); err != nil {
		t.Log("skipping the device tests as ADB is not installed")
		return
	}
	devices, err := adb.Devices()
	if err != nil {
		t.Fatal(err)
	}
	switch len(devices) {
	case 0:
		t.Log("skipping the device tests as none was found via ADB")
		return
	case 1:
		// continue below
	default:
		t.Log("skipping the device tests as too many were found via ADB")
		return
	}

	t.Run("DevicesOne", func(t *testing.T) {
		mustSucceed(t, `\n`, ``, cmdDevices)
	})

	// try to uninstall the app first
	devices[0].Uninstall(chosenApp)
	t.Run("UninstallMissing", func(t *testing.T) {
		mustFail(t, `not installed$`, ``, cmdUninstall, chosenApp)
	})
	t.Run("SearchInstalledMissing", func(t *testing.T) {
		mustSucceed(t, ``, regexp.QuoteMeta(chosenApp), cmdSearch, "-i", "-q")
	})
	t.Run("SearchUpgradableMissing", func(t *testing.T) {
		mustSucceed(t, ``, regexp.QuoteMeta(chosenApp), cmdSearch, "-u", "-q")
	})
	t.Run("InstallVersioned", func(t *testing.T) {
		mustSucceed(t, `Installing `+regexp.QuoteMeta(chosenApp), ``,
			cmdInstall, chosenApp+":1")
	})
	t.Run("SearchInstalled", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		mustSucceed(t, regexp.QuoteMeta(chosenApp), ``, cmdSearch, "-i", "-q")
	})
	t.Run("SearchUpgradable", func(t *testing.T) {
		mustSucceed(t, regexp.QuoteMeta(chosenApp), ``, cmdSearch, "-u", "-q")
	})
	t.Run("InstallUpgrade", func(t *testing.T) {
		mustSucceed(t, `Installing `+regexp.QuoteMeta(chosenApp), ``,
			cmdInstall, chosenApp)
	})
	t.Run("SearchUpgradableUpToDate", func(t *testing.T) {
		mustSucceed(t, ``, regexp.QuoteMeta(chosenApp), cmdSearch, "-u", "-q")
	})
	t.Run("InstallUpToDate", func(t *testing.T) {
		mustSucceed(t, `is up to date$`, ``, cmdInstall, chosenApp)
	})
	t.Run("UninstallExisting", func(t *testing.T) {
		mustSucceed(t, `Uninstalling `+regexp.QuoteMeta(chosenApp), ``,
			cmdUninstall, chosenApp)
	})
}

func mustRun(t *testing.T, success bool, wantRe, negRe string, cmd *Command, args ...string) {
	var buf bytes.Buffer
	err := cmd.Run(args)
	out := buf.String()
	if err != nil {
		out += err.Error()
	}
	if success && err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	} else if !success && err == nil {
		t.Fatalf("expected error, got none\n%s", out)
	}
	// Let '.' match newlines, and treat the output as a single line.
	wantRe = "(?sm)" + wantRe
	if !regexp.MustCompile(wantRe).MatchString(out) {
		t.Fatalf("output does not match %#q:\n%s", wantRe, out)
	}
	if negRe != "" {
		negRe = "(?sm)" + negRe
		if regexp.MustCompile(negRe).MatchString(out) {
			t.Fatalf("output does match %#q:\n%s", negRe, out)
		}
	}
}
