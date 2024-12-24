// Copyright (c) 2024, Thomas Dickson
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

var cmdSetup = &Command{
	UsageLine: "setup",
	Short:     "Manage setups",
	Long: ` 
List, add, remove, edit, and import/export setups.
Setups allow for mass installs onto an android device, excellent for backups.

List setups:

	$ fdroidcl setup                Show all setups
	$ fdroidcl setup list <NAME>    Show details about one setup

Modify setups:

	$ fdroidcl setup new <NAME>
	$ fdroidcl setup remove <NAME>
	$ fdroidcl setup apply <NAME>
	$ fdroidcl setup add-app <NAME> <APP-ID>
	$ fdroidcl setup rm-app <NAME> <APP-ID>
	$ fdroidcl setup add-repo <NAME> <REPO-NAME>
	$ fdroidcl setup rm-repo <NAME> <REPO-NAME>

Export setups:

	$ fdroidcl setup import <FILENAME>
	$ fdroidcl setup export <NAME>
`[1:],
}

func init() {
	cmdSetup.Run = runSetup
}

func runSetup(args []string) error {
	if len(args) == 0 {
		// list repositories
		if len(config.Setups) == 0 {
			fmt.Println("No setups!")
		} else {
			for _, value := range config.Setups {
				fmt.Printf("Name: %s\n", value.ID)
			}
		}
		return nil
	}

	switch args[0] {
	case "list":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return listSetup(args[1])
	case "new":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return newSetup(args[1])
	case "apply":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return applySetup(args[1])
	case "remove":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return removeSetup(args[1])
	case "add-app":
		if len(args) != 3 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return addSetupApp(args[1], args[2])
	case "rm-app":
		if len(args) != 3 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return removeSetupApp(args[1], args[2])
	case "add-repo":
		if len(args) != 3 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return removeSetupRepo(args[1], args[2])
	case "rm-repo":
		if len(args) != 3 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return removeSetupRepo(args[1], args[2])
	case "import":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return importSetup(args[1])
	case "export":
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return exportSetup(args[1])

	}
	return fmt.Errorf("wrong usage")
}

func getIndex(sl []string, str string) int {
	index := -1
	for i, val := range sl {
		if val == str {
			index = i
			break
		}
	}
	return index
}

func listSetup(name string) error {
	index := setupIndex(name)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", name)
	}

	setup := config.Setups[index]
	fmt.Printf("Name: %s\n", setup.ID)
	fmt.Printf("Repos: %s\n", setup.Repos)
	fmt.Printf("Apps: %s\n", setup.Apps)

	return nil
}

func setupIndex(name string) int {
	index := -1
	for i, value := range config.Setups {
		if value.ID == name {
			index = i
			break
		}
	}
	return index
}

func newSetup(name string) error {
	repos := []string{}
	for _, repo := range config.Repos {
		repos = append(repos, repo.ID)
	}
	config.Setups = append(config.Setups, setup{name, []string{}, repos})
	return writeConfig(&config)
}

func applySetup(name string) error {
	index := setupIndex(name)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", name)
	}

	setup := config.Setups[index]
	for _, repo := range setup.Repos {
		index = repoIndex(repo)
		if index == -1 {
			return fmt.Errorf("setup contains unknown repo id \"%s\" ", repo)
		}
	}
	fmt.Println("All repos work!")

	if len(setup.Apps) == 0 {
		return fmt.Errorf("setup has no apps!")
	}

	for _, app := range setup.Apps {
		err := runInstall([]string{app})
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func removeSetup(name string) error {
	index := setupIndex(name)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", name)
	}
	config.Setups = append(config.Setups[:index], config.Setups[index+1:]...)
	return writeConfig(&config)
}

func addSetupApp(setupName string, appName string) error {
	index := setupIndex(setupName)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", setupName)
	}
	config.Setups[index].Apps = append(config.Setups[index].Apps, appName)
	return writeConfig(&config)
}

func removeSetupApp(setupName string, appName string) error {
	index := setupIndex(setupName)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", setupName)
	}

	apps := config.Setups[index].Apps
	appIndex := getIndex(apps, appName)
	if appIndex == -1 {
		return fmt.Errorf("a app with the name \"%s\" could not be found for the setup with the name \"%s\"", appName, setupName)
	}
	config.Setups[index].Apps = append(apps[:appIndex], apps[appIndex+1:]...)

	return writeConfig(&config)
}

func addSetupRepo(setupName string, repoName string) error {
	repoIndex := repoIndex(repoName)
	if repoIndex == -1 {
		return fmt.Errorf("a repo with the name \"%s\" could not be found", repoName)
	}

	index := setupIndex(setupName)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", setupName)
	}
	config.Setups[index].Repos = append(config.Setups[index].Repos, repoName)
	return writeConfig(&config)
}

func removeSetupRepo(setupName string, repoName string) error {
	index := setupIndex(setupName)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", setupName)
	}

	repos := config.Setups[index].Repos
	repoIndex := getIndex(repos, repoName)
	if repoIndex == -1 {
		return fmt.Errorf("a repo with the name \"%s\" could not be found for the setup with the name \"%s\"", repoName, setupName)
	}
	config.Setups[index].Repos = append(repos[:repoIndex], repos[repoIndex+1:]...)

	return writeConfig(&config)
}

func importSetup(filename string) error {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("file \"%s\" does not exist", filename)
	}
	defer f.Close()
	fileSetup := setup{}
	err = toml.NewDecoder(f).Decode(&fileSetup)
	if err != nil {
		return err
	}

	config.Setups = append(config.Setups, fileSetup)
	return writeConfig(&config)
}

func exportSetup(name string) error {
	index := setupIndex(name)
	if index == -1 {
		return fmt.Errorf("a setup with the name \"%s\" could not be found", name)
	}
	b, err := toml.Marshal(config.Setups[index])
	if err != nil {
		return fmt.Errorf("cannot encode config: %v", err)
	}
	f, err := os.Create(config.Setups[index].ID + ".toml")
	if err != nil {
		return fmt.Errorf("cannot create config file: %v", err)
	}
	_, err = f.Write(b)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
