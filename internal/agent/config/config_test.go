package config

import (
	"os"
	"testing"

	"github.com/sotavant/yandex-metrics/internal"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ReadConfigFromFile(t *testing.T) {
	internal.InitLogger()
	configStr := `
{
  "address": "localhost:3456",
  "report_interval": "133s",
  "crypto_key": "somePath",
  "poll_interval": "122s"
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
				Addr:           "localhost:3456",
				ReportInterval: 133,
				PollInterval:   122,
				CryptoKeyPath:  "somePath",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setConf()
			InitConfig()
			assert.Equal(t, tt.want.Addr, AppConfig.Addr)
			assert.Equal(t, tt.want.ReportInterval, AppConfig.ReportInterval)
			assert.Equal(t, tt.want.CryptoKeyPath, AppConfig.CryptoKeyPath)
			assert.Equal(t, tt.want.PollInterval, AppConfig.PollInterval)
		})
	}
}
