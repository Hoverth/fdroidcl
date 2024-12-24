// Copyright (c) 2024, Thomas Dickson
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"strings"

	"mvdan.cc/fdroidcl/adb"
)

var cmdScan = &Command{
	UsageLine: "scan",
	Short:     "Scan for all recognised apps on a connected device",
}

func init() {
	cmdScan.Run = runScan
}

func runScan(args []string) error {
	if err := startAdbIfNeeded(); err != nil {
		return err
	}
	devices, err := adb.Devices()
	if err != nil {
		return fmt.Errorf("could not get devices: %v", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no devices found")
	}

	for _, device := range devices {
		fmt.Fprintf(os.Stderr, "Scanning %s - %s (%s)\n", device.ID, device.Model, device.Product)
		scanForPackages(device)
		fmt.Fprintln(os.Stderr, "Scan completed without error")
	}
	return nil
}

func scanForPackages(device *adb.Device) {
	cmd := device.AdbShell("pm list packages")

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	// fmt.Println(string(out))
	// otherwise, print the output from running the command
	lines := strings.Split(string(out), "\n")
	// fmt.Println(lines)

	for _, line := range lines {
		if len(line) > 8 {
			line = line[8:]
			// fmt.Println(line)

			apps, err := findApps([]string{line})
			if err == nil {
				fmt.Println(apps[0].PackageName)
			}
		}
	}
}
