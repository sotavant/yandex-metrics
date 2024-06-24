// Package config Данный пакет предназначен для считывания и установки параметров
// необходимых для работы приложения.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/sotavant/yandex-metrics/internal"
)

// Параметры по-умолчанию
const (
	DefaultServerAddress = "localhost:8080"
	DefaultTableName     = "metric"
	DefaultStoreInterval = 300
	DefaultMetricDB      = "/tmp/metrics-db.json"
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
	CryptKeyVar        = `CRYPTO_KEY`
	configPathKeyVar   = `CONFIG`
)

// fileConfig для настроек из файла конфига
type fileConfig struct {
	Address          string `json:"address"`
	StoreIntervalStr string `json:"store_interval"`
	StoreFile        string `json:"store_file"`
	DatabaseDSN      string `json:"database_dsn"`
	CryptoKey        string `json:"crypto_key"`
	Restore          bool   `json:"restore"`
}

// Config Структура для хранения параметров
type Config struct {
	Addr            string `json:"address"`
	HashKey         string
	FileStoragePath string `json:"store_file"`
	DatabaseDSN     string `json:"database_dsn"`
	TableName       string
	CryptoKeyPath   string `json:"crypto_key"`
	StoreInterval   uint
	Restore         bool `json:"restore"`
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
	var address, storeFile, databaseDsn, cryptoKey, config, cnfShort string
	var restore bool
	var storeInterval uint

	flag.StringVar(&address, "a", "", "server address")
	flag.BoolVar(&restore, "r", true, "need restore values")
	flag.UintVar(&storeInterval, "i", 0, "store interval")
	flag.StringVar(&storeFile, "f", "", "file storage path")
	flag.StringVar(&databaseDsn, "d", "", "database DSN")
	flag.StringVar(&c.TableName, "t", DefaultTableName, "table name")
	flag.StringVar(&c.HashKey, "k", "dd", "hash key")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to public key")
	flag.StringVar(&config, "config", "", "path to config file")
	flag.StringVar(&cnfShort, "c", "", "path to config file")

	if config == "" {
		config = cnfShort
	}

	flag.Parse()

	c.readConfig(config)

	if address != "" {
		c.Addr = address
	} else if c.Addr == "" {
		c.Addr = DefaultServerAddress
	}

	if !restore {
		c.Restore = restore
	}

	if storeInterval == 0 {
		if c.StoreInterval == 0 {
			c.StoreInterval = DefaultStoreInterval
		}
	} else {
		c.StoreInterval = storeInterval
	}

	if storeFile != "" {
		c.FileStoragePath = storeFile
	} else if c.FileStoragePath == "" {
		c.FileStoragePath = DefaultMetricDB
	}

	if databaseDsn != "" {
		c.DatabaseDSN = databaseDsn
	}

	if cryptoKey != "" {
		c.CryptoKeyPath = cryptoKey
	}

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

	if cryptoKey := os.Getenv(CryptKeyVar); cryptoKey != "" {
		c.CryptoKeyPath = cryptoKey
	}
}

// readFile чтение конфигурации из файла
func (c *Config) readConfig(configPath string) {
	if configPathEnv := os.Getenv(configPathKeyVar); configPathEnv != "" {
		configPath = configPathEnv
	}

	if configPath == "" {
		return
	}

	cnf, err := os.ReadFile(configPath)
	if err != nil {
		internal.Logger.Fatalw("failed to read config file", "err", err)
		panic(err)
	}

	fileCnf := &fileConfig{}
	err = json.Unmarshal(cnf, fileCnf)
	if err != nil {
		internal.Logger.Fatalw("failed to parse config file", "err", err)
		panic(err)
	}

	if fileCnf.StoreIntervalStr != "" {
		var interval uint64
		interval, err = strconv.ParseUint(strings.TrimSuffix(fileCnf.StoreIntervalStr, "s"), 10, 32)
		if err != nil {
			panic(err)
		}

		c.StoreInterval = uint(interval)
	}

	if fileCnf.CryptoKey != "" {
		c.CryptoKeyPath = fileCnf.CryptoKey
	}

	if fileCnf.Address != "" {
		c.Addr = fileCnf.Address
	}

	c.Restore = fileCnf.Restore

	if fileCnf.DatabaseDSN != "" {
		c.DatabaseDSN = fileCnf.DatabaseDSN
	}

	if fileCnf.StoreFile != "" {
		c.FileStoragePath = fileCnf.StoreFile
	}
}
