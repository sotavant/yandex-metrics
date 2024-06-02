// Package config Данный пакет предназначен для считывания и установки параметров
// необходимых для работы приложения.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Параметры по-умолчанию
const (
	DefaultServerAddress = "localhost:8080"
	DefaultTableName     = "metric"
)

// Названия переменных окружения
const (
	addressVar         = `ADDRESS`
	storeIntervalVar   = `STORE_INTERVAL`
	fileStoragePathVar = `FILE_STORAGE_PATH`
	restoreVar         = `RESTORE`
	databaseDSNVar     = `DATABASE_DSN`
	tableNameVar       = `TABLE_NAME`
	HashKeyVar         = `KEY`
)

// Config Структура для хранения параметров
type Config struct {
	Addr            string
	HashKey         string
	FileStoragePath string
	DatabaseDSN     string
	TableName       string
	StoreInterval   uint
	Restore         bool
}

// InitConfig инициализация конфигурации
func InitConfig() *Config {
	conf := new(Config)
	conf.ParseFlags()
	return conf
}

// ParseFlags Метод для считывания параметров.
// Сначала считываются значения из командной строки, если они не заданы, то берутся значения по-умолчанию
// Если заданы переменные окружения, то они переопределяют значения заданные ранее
func (c *Config) ParseFlags() {
	flag.StringVar(&c.Addr, "a", DefaultServerAddress, "server address")
	flag.BoolVar(&c.Restore, "r", true, "need restore values")
	flag.UintVar(&c.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&c.TableName, "t", DefaultTableName, "table name")
	flag.StringVar(&c.HashKey, "k", "", "hash key")

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

	if hashKey := os.Getenv(HashKeyVar); hashKey != "" {
		c.HashKey = hashKey
	}
}
