package config

import (
	"os"
	"testing"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
)

func TestServerConfig_ReadConfigFromFile(t *testing.T) {
	internal.InitLogger()
	configStr := `
{
    "address": "localhost:8083",
    "restore": false,
    "store_interval": "1s",
    "store_file": "/path/to/file.db",
    "database_dsn": "",
    "crypto_key": "/path/to/key.pem" 
} 
`
	file, err := os.CreateTemp(os.TempDir(), "config")
	assert.NoError(t, err)

	err = os.WriteFile(file.Name(), []byte(configStr), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		setConf func()
		want    Config
	}{
		{
			name: "With fileConfig",
			setConf: func() {
				err = os.Setenv(configPathKeyVar, file.Name())
				assert.NoError(t, err)
			},
			want: Config{
				Addr:            "localhost:8083",
				Restore:         false,
				StoreInterval:   1,
				FileStoragePath: "/path/to/file.db",
				DatabaseDSN:     "",
				CryptoKeyPath:   "/path/to/key.pem",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setConf()
			conf := InitConfig()
			assert.Equal(t, tt.want.Addr, conf.Addr)
			assert.Equal(t, tt.want.Restore, conf.Restore)
			assert.Equal(t, tt.want.StoreInterval, conf.StoreInterval)
			assert.Equal(t, tt.want.FileStoragePath, conf.FileStoragePath)
			assert.Equal(t, tt.want.DatabaseDSN, conf.DatabaseDSN)
			assert.Equal(t, tt.want.CryptoKeyPath, conf.CryptoKeyPath)
		})
	}
}
