# fdroidcl

[![GoDoc](https://godoc.org/github.com/mvdan/fdroidcl?status.svg)](https://godoc.org/mvdan.cc/fdroidcl)

[F-Droid](https://f-droid.org/) desktop client. Requires Go 1.19 or later.

```sh
 go install mvdan.cc/fdroidcl@latest
```

While the Android client integrates with the system with regular update checks
and notifications, this is a simple command line client that talks to connected
devices via [ADB](https://developer.android.com/tools/help/adb.html).

## Quickstart

Download the index:

```sh
 fdroidcl update
```

Show all available apps:

```sh
 fdroidcl search
```

Install an app:

```sh
 fdroidcl install org.adaway
```

Show all available updates, and install them:

```sh
 fdroidcl search -u
 fdroidcl install -u
```

Unofficial packages are available on:
[Debian](https://packages.debian.org/buster/fdroidcl) and
[Ubuntu](https://packages.ubuntu.com/eoan/fdroidcl).

## Commands

```text
 update                   Update the index
 search [<regexp...>]     Search available apps
 show <appid...>          Show detailed info about apps
 install [<appid...>]     Install or upgrade apps
 uninstall <appid...>     Uninstall an app
 download <appid...>      Download an app
 devices                  List connected devices
 scan                     Scan for known fdroid apps on a device
 list (categories/users)  List all known values of a kind
 repo                     Manage repositories
 setups                   Manage setups
 clean                    Clean index and/or cache
 defaults                 Reset to the default settings
 version                  Print version information
```

An appid is just an app's unique package name. A specific version of an app can
be selected by following the appid with a colon and the version code. The
'search' and 'show' commands can be used to find these strings. For example:

```sh
 fdroidcl search redreader
 fdroidcl show org.quantumbadger.redreader
 fdroidcl install org.quantumbadger.redreader:85
```

### *new: you can manage the repositories now directly via cli*

```text
usage: fdroidcl repo

List, add, remove, enable or disable repositories.
When a repository is added, it is enabled by default.

List repositories:

        $ fdroidcl repo

Modify repositories:

        $ fdroidcl repo add <NAME> <URL>
        $ fdroidcl repo remove <NAME>
        $ fdroidcl repo enable <NAME>
        $ fdroidcl repo disable <NAME>
```

### *new: you can manage automating installs using setups*

```text
usage: fdroidcl setup

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
```

## Config

You can configure what repositories to use in the `config.toml` file. On Linux,
you will likely find it at `~/.config/fdroidcl/config.toml`.

You can run `fdroidcl defaults` to create the config with the default settings.

## Advantages over the Android client

* Command line interface
* Batch install/update/remove apps without root nor system privileges
* No need to install a client on the device

## What it will never do

* Run as a daemon, e.g. periodic index updates
* Act as an F-Droid server
* Swap apps with devices

## Caveats

* Index verification relies on HTTPS (not the JAR signature)
* The tool can only interact with one device at a time
* Hardware compatibility of packages is not checked

## FAQ

* What's the point of a desktop client?

This client works with Android devices connected via ADB; it does not install
apps on the host machine.

* Why not just use the f-droid.org website to download APKs?

That's always an option. However, an F-Droid client supports multiple
repositories, searching for apps, filtering by compatibility with your device,
showing available updates, et cetera.
