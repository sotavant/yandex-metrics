// Package config Данный пакет служит для установки параметров.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/sotavant/yandex-metrics/internal"
)

// параметы по-умолчанию.
const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	rateLimit             = 10
	serverAddress         = `localhost:8080`
)

// названия переменных окружения.
const (
	addressVar       = `ADDRESS`
	reportIntVar     = `REPORT_INTERVAL`
	pollIntVar       = `POLL_INTERVAL`
	HashKeyVar       = `KEY`
	RateLimitVar     = `RATE_LIMIT`
	CryptKeyVar      = `CRYPTO_KEY`
	CryptCertVar     = `CRYPTO_CERT`
	configPathKeyVar = `CONFIG`
)

// AppConfig глобальная переменная, в которой хранятся конфигурации.
var AppConfig *Config

// fileConfig для настроек из файла конфига
type fileConfig struct {
	Address           string `json:"address"`
	PollIntervalStr   string `json:"poll_interval"`
	ReportIntervalStr string `json:"report_interval"`
	CryptoKey         string `json:"crypto_key"`
	CryptoCert        string `json:"crypto_cert"`
}

// Config структура для хранения настроек.
type Config struct {
	Addr           string
	HashKey        string
	CryptoKeyPath  string
	CryptoCertPath string
	ReportInterval int
	PollInterval   int
	RateLimit      int
}

// InitConfig инициализация значения конфигурации.
func InitConfig() {
	AppConfig = new(Config)
	AppConfig.ParseFlags()
}

// ParseFlags считыванание значений либо из параметров запуска либо из переменных окружения
func (c *Config) ParseFlags() {
	var address, cryptoKey, cryptoCert, config, cnfShort string
	var pullInterval, reportIntervalFlag int

	flag.StringVar(&address, "a", "", "server address")
	flag.StringVar(&cryptoKey, "crypto-key", "", "path to public key")
	flag.StringVar(&cryptoCert, "crypto-cert", "", "path to cert key")
	flag.IntVar(&pullInterval, "p", 0, "pollInterval")
	flag.IntVar(&reportIntervalFlag, "r", 0, "reportInterval")
	flag.StringVar(&c.HashKey, "k", "dd", "hash key")
	flag.IntVar(&c.RateLimit, "l", rateLimit, "rate limit")
	flag.StringVar(&config, "config", "", "path to config file")
	flag.StringVar(&cnfShort, "c", "", "path to config file")

	flag.Parse()

	if config == "" {
		config = cnfShort
	}

	c.readConfig(config)

	if address != "" {
		c.Addr = address
	}

	if cryptoKey != "" {
		c.CryptoKeyPath = cryptoKey
	}

	if cryptoCert != "" {
		c.CryptoCertPath = cryptoCert
	}

	if pullInterval != 0 {
		c.PollInterval = pullInterval
	} else if c.PollInterval == 0 {
		c.PollInterval = defaultPollInterval
	}

	if reportIntervalFlag != 0 {
		c.ReportInterval = reportIntervalFlag
	} else if c.ReportInterval == 0 {
		c.ReportInterval = defaultReportInterval
	}

	if envAddr := os.Getenv(addressVar); envAddr != "" {
		c.Addr = envAddr
	}

	if envReport := os.Getenv(reportIntVar); envReport != "" {
		reportIntervalEnvVal, err := strconv.Atoi(envReport)
		if err == nil {
			c.ReportInterval = reportIntervalEnvVal
		}
	}

	if envPoll := os.Getenv(pollIntVar); envPoll != "" {
		pollIntervalEnvVal, err := strconv.Atoi(envPoll)
		if err == nil {
			c.PollInterval = pollIntervalEnvVal
		} else {
			internal.Logger.Infow("poll interval convert error", "err", err)
		}
	}

	if rateLimitEnv := os.Getenv(RateLimitVar); rateLimitEnv != "" {
		rateLimitEnvVal, err := strconv.Atoi(rateLimitEnv)
		if err == nil {
			c.RateLimit = rateLimitEnvVal
		} else {
			internal.Logger.Infow("rate limit convert error", "err", err)
		}
	}

	if hashKey := os.Getenv(HashKeyVar); hashKey != "" {
		c.HashKey = hashKey
	}

	if cryptoKeyEnv := os.Getenv(CryptKeyVar); cryptoKeyEnv != "" {
		c.CryptoKeyPath = cryptoKeyEnv
	}

	if cryptoCertEnv := os.Getenv(CryptCertVar); cryptoCertEnv != "" {
		c.CryptoCertPath = cryptoCertEnv
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
	}

	fileCnf := &fileConfig{}
	err = json.Unmarshal(cnf, fileCnf)
	if err != nil {
		internal.Logger.Fatalw("failed to parse config file", "err", err)
	}

	if fileCnf.ReportIntervalStr != "" {
		c.ReportInterval, err = strconv.Atoi(strings.TrimSuffix(fileCnf.ReportIntervalStr, "s"))
		if err != nil {
			internal.Logger.Fatalw("failed to parse report interval from file", "err", err)
		}
	}

	if fileCnf.PollIntervalStr != "" {
		c.PollInterval, err = strconv.Atoi(strings.TrimSuffix(fileCnf.PollIntervalStr, "s"))
		if err != nil {
			internal.Logger.Fatalw("failed to parse poll interval from file", "err", err)
		}
	}

	if fileCnf.CryptoKey != "" {
		c.CryptoKeyPath = fileCnf.CryptoKey
	}

	if fileCnf.CryptoCert != "" {
		c.CryptoCertPath = fileCnf.CryptoCert
	}

	if fileCnf.Address != "" {
		c.Addr = fileCnf.Address
	}
}
