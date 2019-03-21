package main

import (
	"flag"
	"fmt"
	"greenskeeper"
	"os"
	"os/user"
	"strconv"
)

func main() {
	var rootlessMode bool
	flag.BoolVar(&rootlessMode, "rootless", false, "run rootless setup")

	flag.Parse()

	owner := 0
	if rootlessMode {
		owner = mustGetMaximus()
	}

	pidFilePath := os.Getenv("PIDFILE")
	if err := greenskeeper.CheckExistingGdnProcess(pidFilePath); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	directories := []greenskeeper.Directory{
		greenskeeper.NewDirectoryBuilder(mustGetenv("RUN_DIR")).Mode(0770).Build(),
		greenskeeper.NewDirectoryBuilder(mustGetenv("GARDEN_DATA_DIR")).Mode(0770).UID(mustResolveUID("vcap")).GID(mustGetMaximus()).Build(),
		greenskeeper.NewDirectoryBuilder(mustGetenv("LOG_DIR")).Mode(0770).UID(owner).GID(owner).Build(),
		greenskeeper.NewDirectoryBuilder(mustGetenv("TMPDIR")).Mode(0755).UID(owner).GID(owner).Build(),
		greenskeeper.NewDirectoryBuilder(mustGetenv("DEPOT_PATH")).Mode(0755).UID(owner).GID(owner).Build(),
		greenskeeper.NewDirectoryBuilder(mustGetenv("RUNTIME_BIN_DIR")).Mode(0750).GID(mustGetMaximus()).Build(),
	}

	if rootlessMode {
		directories = append(directories, greenskeeper.NewDirectoryBuilder(mustGetenv("XDG_RUNTIME_DIR")).Mode(0700).UID(owner).GID(owner).Build())
		directories = append(directories, greenskeeper.NewDirectoryBuilder(mustGetenv("GARDEN_ROOTLESS_CONFIG_DIR")).Mode(0700).UID(owner).GID(owner).Build())
		directories = append(directories, greenskeeper.NewDirectoryBuilder(mustGetenv("CONTAINERD_DATA_DIR")).Mode(0700).UID(owner).GID(owner).Build())
	}

	if err := greenskeeper.CreateDirectories(directories...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func mustGetenv(key string) string {
	env := os.Getenv(key)
	if env == "" {
		fmt.Fprintf(os.Stderr, "expected environment variable %s to be set", key)
		os.Exit(1)
	}

	return env
}

func mustGetMaximus() int {
	maximus := mustGetenv("MAXIMUS")

	maximusID, err := strconv.Atoi(maximus)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error converting '%s' string to int", maximus)
		os.Exit(1)
	}

	return maximusID
}

func mustResolveUID(username string) int {
	u, err := user.Lookup(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "expected user %s to exsit", username)
		os.Exit(1)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not convert user %s to UID", username)
		os.Exit(1)
	}

	return uid
}
