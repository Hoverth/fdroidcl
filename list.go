// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var cmdList = &Command{
	UsageLine: "list (categories/users/repos)",
	Short:     "List all known values of a kind",
}

func init() {
	cmdList.Run = runList
}

func runList(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need exactly one argument")
	}
	apps, err := loadIndexes()
	if err != nil {
		return err
	}
	values := make(map[string]struct{})
	switch args[0] {
	case "categories":
		for _, app := range apps {
			for _, c := range app.Categories {
				values[c] = struct{}{}
			}
		}
	case "users":
		if err := listUsers(); err != nil {
			return err
		}
	case "repos":
		for _, app := range apps {
			values[app.FdroidRepoName] = struct{}{}
		}
	default:
		return fmt.Errorf("invalid argument")
	}
	result := make([]string, 0, len(values))
	for s := range values {
		result = append(result, s)
	}
	sort.Strings(result)
	for _, s := range result {
		fmt.Fprintln(os.Stdout, s)
	}
	return nil
}

var userIdNameRegex = regexp.MustCompile(`UserInfo{(\d+):([^:}]*):[^}]*}`)

func listUsers() error {
	device, err := oneDevice()
	if err != nil {
		return err
	}
	cmd := device.AdbShell("pm", "list", "users")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	uidHeader := "UID"
	nameHeader := "Name"
	runningHeader := "Running"
	scanner := bufio.NewScanner(stdout)
	uids := make([]string, 0)
	names := make([]string, 0)
	running := make([]bool, 0)
	maxUidLen := len(uidHeader)
	maxNameLen := len(nameHeader)
	for scanner.Scan() {
		text := scanner.Text()
		m := userIdNameRegex.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		uid := m[1]
		uids = append(uids, uid)
		if uidLen := len(uid); uidLen > maxUidLen {
			maxUidLen = uidLen
		}
		name := m[2]
		names = append(names, name)
		if nameLen := len(name); nameLen > maxNameLen {
			maxNameLen = nameLen
		}
		currentRunning := false
		if strings.HasSuffix(strings.TrimSpace(text), "running") {
			currentRunning = true
		}
		running = append(running, currentRunning)
	}
	if len(uids) == 0 {
		return nil
	}
	fmt.Printf("%-*s %-*s %s\n", maxUidLen, uidHeader, maxNameLen, nameHeader, runningHeader)
	for i, uid := range uids {
		name := names[i]
		runningStr := ""
		if running[i] {
			runningStr = "Yes"
		}
		fmt.Printf("%*s %-*s %s\n", maxUidLen, uid, maxNameLen, name, runningStr)
	}
	return nil
}
