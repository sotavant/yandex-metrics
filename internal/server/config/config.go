package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	addressVar           = `ADDRESS`
	storeIntervalVar     = `STORE_INTERVAL`
	fileStoragePathVar   = `FILE_STORAGE_PATH`
	restoreVar           = `RESTORE`
	databaseDSNVar       = `DATABASE_DSN`
	tableNameVar         = `TABLE_NAME`
	DefaultServerAddress = "localhost:8080"
	DefaultTableName     = "metric"
)

type Config struct {
	Addr            string
	StoreInterval   uint
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	TableName       string
}

func InitConfig() *Config {
	conf := new(Config)
	conf.ParseFlags()
	return conf
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Addr, "a", DefaultServerAddress, "server address")
	flag.BoolVar(&c.Restore, "r", true, "need restore values")
	flag.UintVar(&c.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&c.TableName, "t", DefaultTableName, "table name")

	flag.Parse()

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.Addr = envAddr
	}

	if storeInt := os.Getenv(storeIntervalVar); storeInt != "" {
		intVal, err := strconv.ParseUint(storeInt, 10, 32)
		if err != nil {
			panic(err)
		}

		c.StoreInterval = uint(intVal)
	}

	if storageFile := os.Getenv(fileStoragePathVar); storageFile != "" {
		c.FileStoragePath = storageFile
	}

	if needRestore := os.Getenv(restoreVar); needRestore != "" {
		boolVal, err := strconv.ParseBool(needRestore)
		if err != nil {
			panic(err)
		}

		c.Restore = boolVal
	}

	if databaseDSN := os.Getenv(databaseDSNVar); databaseDSN != "" {
		c.DatabaseDSN = databaseDSN
	}

	if tblName := os.Getenv(tableNameVar); tblName != "" {
		c.TableName = tblName
	}
}
