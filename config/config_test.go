package config_test

import (
	"testing"

	"github.com/sixexorg/magnetic-ring/config"
)

func TestDecodeConfig(t *testing.T) {
	configPath := "./config.yml"
	err := config.InitConfig(configPath)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("config:%+v\n", config.GlobalConfig)
	for k, v := range config.GlobalConfig.Genesis.Stars {
		t.Log(k, v.crystal, v.Account, v.Nodekey)
	}

}
