package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	addressVar         = `ADDRESS`
	storeIntervalVar   = `STORE_INTERVAL`
	fileStoragePathVar = `FILE_STORAGE_PATH`
	restoreVar         = `RESTORE`
)

type config struct {
	addr            string
	storeInterval   uint
	fileStoragePath string
	restore         bool
}

func (c *config) parseFlags() {
	flag.StringVar(&c.addr, "a", serverAddress, "server address")
	flag.BoolVar(&c.restore, "r", true, "need restore values")
	flag.UintVar(&c.storeInterval, "i", 300, "store interval")
	flag.StringVar(&c.fileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")

	flag.Parse()

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.addr = envAddr
	}

	if storeInt := os.Getenv(storeIntervalVar); storeInt != "" {
		intVal, err := strconv.ParseUint(storeInt, 10, 32)
		if err != nil {
			panic(err)
		}

		c.storeInterval = uint(intVal)
	}

	if storageFile := os.Getenv(fileStoragePathVar); storageFile != "" {
		c.fileStoragePath = storageFile
	}

	if needRestore := os.Getenv(restoreVar); needRestore != "" {
		boolVal, err := strconv.ParseBool(needRestore)
		if err != nil {
			panic(err)
		}

		c.restore = boolVal
	}
}
