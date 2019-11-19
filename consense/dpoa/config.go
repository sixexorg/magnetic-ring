package dpoa

import (
	"github.com/sixexorg/magnetic-ring/account"
)

type EartchCfg struct {
	duration int //10
}

type StarsCfg struct {
	duration int //10
}

type Config struct {
	accountStr  string
	account     account.NormalAccount
	Peers       []string `json:"peers"`
	earthCfg    *EartchCfg
	starsCfg     *StarsCfg
}


var DftConfig *Config = &Config{Peers:make([]string, 0), earthCfg:&EartchCfg{duration:10}, starsCfg:&StarsCfg{duration:10}}
