// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"mvdan.cc/fdroidcl/adb"
	"mvdan.cc/fdroidcl/fdroid"
)

var cmdSearch = &Command{
	UsageLine: "search [<regexp...>]",
	Short:     "Search available apps",
}

var (
	searchQuiet     = cmdSearch.Fset.Bool("q", false, "Print package names only")
	searchInstalled = cmdSearch.Fset.Bool("i", false, "Filter installed apps")
	searchUpdates   = cmdSearch.Fset.Bool("u", false, "Filter apps with updates")
	searchDays      = cmdSearch.Fset.Int("d", 0, "Select apps last updated in the last <n> days; a negative value drops them instead")
	searchCategory  = cmdSearch.Fset.String("c", "", "Filter apps by category")
	searchRepo      = cmdSearch.Fset.String("r", "", "Filter apps by repository")
	searchSortBy    = cmdSearch.Fset.String("o", "", "Sort order (added, updated)")
	searchUser      = cmdSearch.Fset.String("user", "all", "Filter installed apps by user <USER_ID|current|all>")
)

func init() {
	cmdSearch.Run = runSearch
}

func runSearch(args []string) error {
	if *searchInstalled && *searchUpdates {
		return fmt.Errorf("-i is redundant if -u is specified")
	}
	sfunc, err := sortFunc(*searchSortBy)
	if err != nil {
		return err
	}
	apps, err := loadIndexes()
	if err != nil {
		return err
	}
	if len(apps) > 0 {
		if *searchCategory != "" {
			apps = filterAppsCategory(apps, *searchCategory)
			if apps == nil {
				return fmt.Errorf("no such category: %s", *searchCategory)
			}
		}
		if *searchRepo != "" {
			apps = filterAppsRepo(apps, *searchRepo)
			if apps == nil {
				return fmt.Errorf("no such repository: %s", *searchRepo)
			}
		}
	}
	if len(apps) > 0 && len(args) > 0 {
		apps = filterAppsSearch(apps, args)
	}
	var device *adb.Device
	var inst map[string]adb.Package
	if *searchInstalled || *searchUpdates {
		if device, err = oneDevice(); err != nil {
			return err
		}
		if inst, err = device.Installed(); err != nil {
			return err
		}
	}
	var filterUser *int
	if *searchUser != "all" && *searchUser != "current" {
		n, err := strconv.Atoi(*searchUser)
		if err != nil {
			return fmt.Errorf("-user has to be <USER_ID|current|all>")
		}
		if n < 0 {
			return fmt.Errorf("-user cannot have a negative number as USER_ID")
		}
		allUids := adb.AllUserIds(inst)
		if _, exists := allUids[n]; !exists {
			return fmt.Errorf("user %d does not exist", n)
		}
		filterUser = &n
	} else {
		if *installUser == "current" {
			uid, err := device.CurrentUserId()
			if err != nil {
				return err
			}
			filterUser = &uid
		} else {
			filterUser = nil
		}
	}
	if len(apps) > 0 && *searchInstalled {
		apps = filterAppsInstalled(apps, inst, filterUser)
	}
	if len(apps) > 0 && *searchUpdates {
		apps = filterAppsUpdates(apps, inst, device, filterUser)
	}
	if len(apps) > 0 && *searchDays != 0 {
		apps = filterAppsLastUpdated(apps, *searchDays)
	}
	if sfunc != nil {
		apps = sortApps(apps, sfunc)
	}
	if *searchQuiet {
		for _, app := range apps {
			fmt.Fprintln(os.Stdout, app.PackageName)
		}
	} else {
		printApps(apps, inst, device)
	}
	return nil
}

func filterAppsSearch(apps []fdroid.App, terms []string) []fdroid.App {
	regexes := make([]*regexp.Regexp, len(terms))
	for i, term := range terms {
		regexes[i] = regexp.MustCompile(term)
	}
	var result []fdroid.App
	for _, app := range apps {
		fields := []string{
			strings.ToLower(app.PackageName),
			strings.ToLower(app.Name),
			strings.ToLower(app.Summary),
			strings.ToLower(app.Description),
		}
		if !appMatches(fields, regexes) {
			continue
		}
		result = append(result, app)
	}
	return result
}

func appMatches(fields []string, regexes []*regexp.Regexp) bool {
fieldLoop:
	for _, field := range fields {
		for _, regex := range regexes {
			if !regex.MatchString(field) {
				continue fieldLoop
			}
		}
		return true
	}
	return false
}

func printApps(apps []fdroid.App, inst map[string]adb.Package, device *adb.Device) {
	maxIDLen := 0
	for _, app := range apps {
		if len(app.PackageName) > maxIDLen {
			maxIDLen = len(app.PackageName)
		}
	}
	for _, app := range apps {
		var pkg *adb.Package
		p, e := inst[app.PackageName]
		if e {
			pkg = &p
		}
		printApp(app, maxIDLen, pkg, device)
	}
}

func descVersion(app fdroid.App, inst *adb.Package, device *adb.Device) string {
	if inst != nil {
		suggested := app.SuggestedApk(device)
		if suggested != nil && inst.VersCode < suggested.VersCode {
			return fmt.Sprintf("%s (%d) -> %s (%d)", inst.VersName, inst.VersCode,
				suggested.VersName, suggested.VersCode)
		}
		return fmt.Sprintf("%s (%d)", inst.VersName, inst.VersCode)
	}
	return fmt.Sprintf("%s (%d)", app.SugVersName, app.SugVersCode)
}

func printApp(app fdroid.App, IDLen int, inst *adb.Package, device *adb.Device) {
	fmt.Printf("%s%s %s - %s\n", app.PackageName, strings.Repeat(" ", IDLen-len(app.PackageName)),
		app.Name, descVersion(app, inst, device))
	fmt.Printf("    %s\n", app.Summary)
}

func filterAppsInstalled(apps []fdroid.App, inst map[string]adb.Package, user *int) []fdroid.App {
	var result []fdroid.App
	for _, app := range apps {
		p, e := inst[app.PackageName]
		if !e || p.IsSystem {
			continue
		}
		if user != nil {
			installedForUser := false
			for _, appUser := range p.InstalledForUsers {
				if appUser == *user {
					installedForUser = true
					break
				}
			}
			if !installedForUser {
				continue
			}
		}
		result = append(result, app)
	}
	return result
}

func filterAppsUpdates(apps []fdroid.App, inst map[string]adb.Package, device *adb.Device, user *int) []fdroid.App {
	var result []fdroid.App
	for _, app := range apps {
		p, e := inst[app.PackageName]
		if !e || p.IsSystem {
			continue
		}
		if user != nil {
			installedForUser := false
			for _, appUser := range p.InstalledForUsers {
				if appUser == *user {
					installedForUser = true
					break
				}
			}
			if !installedForUser {
				continue
			}
		}
		suggested := app.SuggestedApk(device)
		if suggested == nil {
			continue
		}
		if p.VersCode >= suggested.VersCode {
			continue
		}
		result = append(result, app)
	}
	return result
}

func filterAppsLastUpdated(apps []fdroid.App, days int) []fdroid.App {
	var result []fdroid.App
	newer := true
	if days < 0 {
		days = -days
		newer = false
	}
	date := time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1-days)
	for _, app := range apps {
		if app.Updated.Before(date) == newer {
			continue
		}
		result = append(result, app)
	}
	return result
}

func contains(l []string, s string) bool {
	for _, s1 := range l {
		if s1 == s {
			return true
		}
	}
	return false
}

func filterAppsRepo(apps []fdroid.App, repo string) []fdroid.App {
	var result []fdroid.App
	for _, app := range apps {
		if app.FdroidRepoName != repo {
			continue
		}
		result = append(result, app)
	}
	return result
}

func filterAppsCategory(apps []fdroid.App, categ string) []fdroid.App {
	var result []fdroid.App
	for _, app := range apps {
		if !contains(app.Categories, categ) {
			continue
		}
		result = append(result, app)
	}
	return result
}

func cmpAdded(a, b *fdroid.App) bool {
	return a.Added.Before(b.Added.Time)
}

func cmpUpdated(a, b *fdroid.App) bool {
	return a.Updated.Before(b.Updated.Time)
}

func sortFunc(sortBy string) (func(a, b *fdroid.App) bool, error) {
	switch sortBy {
	case "added":
		return cmpAdded, nil
	case "updated":
		return cmpUpdated, nil
	case "":
		return nil, nil
	}
	return nil, fmt.Errorf("unknown sort order: %s", sortBy)
}

type appList struct {
	l []fdroid.App
	f func(a, b *fdroid.App) bool
}

func (al appList) Len() int           { return len(al.l) }
func (al appList) Swap(i, j int)      { al.l[i], al.l[j] = al.l[j], al.l[i] }
func (al appList) Less(i, j int) bool { return al.f(&al.l[i], &al.l[j]) }

func sortApps(apps []fdroid.App, f func(a, b *fdroid.App) bool) []fdroid.App {
	sort.Sort(appList{l: apps, f: f})
	return apps
}
