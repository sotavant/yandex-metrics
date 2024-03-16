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
	databaseDSNVar     = `DATABASE_DSN`
	tableNameVar       = `TABLE_NAME`
)

type config struct {
	addr            string
	storeInterval   uint
	fileStoragePath string
	restore         bool
	databaseDSN     string
	tableName       string
}

func initConfig() *config {
	conf := new(config)
	conf.parseFlags()
	return conf
}

func (c *config) parseFlags() {
	flag.StringVar(&c.addr, "a", serverAddress, "server address")
	flag.BoolVar(&c.restore, "r", true, "need restore values")
	flag.UintVar(&c.storeInterval, "i", 300, "store interval")
	flag.StringVar(&c.fileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.StringVar(&c.databaseDSN, "d", "", "database DSN")
	flag.StringVar(&c.tableName, "t", TableName, "table name")

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

	if databaseDSN := os.Getenv(databaseDSNVar); databaseDSN != "" {
		c.databaseDSN = databaseDSN
	}

	if tblName := os.Getenv(tableNameVar); tblName != "" {
		c.tableName = tblName
	}
}
