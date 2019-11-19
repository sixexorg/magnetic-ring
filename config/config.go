package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	Version             = "magnetic0.1"
	DefaultKeyDir       = "keystore"
	DefaultMultiAddrDir = "keystore/multiaddrdir"
)

var GlobalConfig MagneticConfig

func InitConfig(configPath string) error {
	var err error
	//path := "./config.yml"
	provide, err := FromConfigString(configPath, "yml")
	if err != nil {
		return err
	}
	GlobalConfig, err = DecodeConfig(provide)
	if err != nil {
		return err
	}
	err = GlobalConfig.CheckConfig([]string{"Genesis", "SysCfg", "TxPoolCfg", "P2PCfg", "CliqueCfg", "BootNode"})
	if err != nil {
		panic(err)
	}
	return nil
}
func (c *MagneticConfig) CheckConfig(cnames []string) error {
	s := reflect.ValueOf(c).Elem()
	for _, v := range cnames {
		val := s.FieldByName(v)
		typ := val.Type()
		def := reflect.New(typ).Elem()
		flag := 0
		i := 0
		for ; i < val.NumField(); i++ {
			if val.Field(i).Kind() != reflect.Slice {
				if val.Field(i).Interface() == def.Field(i).Interface() {
					flag++
				}
			}
		}
		if i == flag {
			return fmt.Errorf("%v is not find in config", v)
		}
	}
	return nil
}

type MagneticConfig struct {
	Genesis   GenesisConfig
	SysCfg    SystemConfig
	TxPoolCfg TxPoolConfig
	P2PCfg    P2PNodeConfig
	CliqueCfg CliqueConfig
	BootNode  BootConfig
	MongoCfg Mongo
}
type BootConfig struct {
	IP string
}
type GenesisConfig struct {
	ChainId   string
	Crystal       uint64
	Energy      uint64
	Official  string
	Timestamp uint64
	Stars     []*StarNode
	//Earth     string
}
type StarNode struct {
	crystal     uint64
	Account string
	Nodekey string
}

type P2PRsvConfig struct {
	ReservedPeers []string `json:"reserved"`
	MaskPeers     []string `json:"mask"`
}

type P2PNodeConfig struct {
	ReservedPeersOnly         bool
	ReservedCfg               *P2PRsvConfig
	NetworkMagic              uint32
	NetworkId                 uint32
	NetworkName               string
	NodePort                  uint
	NodeConsensusPort         uint
	DualPortSupport           bool
	IsTLS                     bool
	CertPath                  string
	KeyPath                   string
	CAPath                    string
	HttpInfoPort              uint
	MaxHdrSyncReqs            uint
	MaxConnInBound            uint
	MaxConnOutBound           uint
	MaxConnInBoundForSingleIP uint
}

type SystemConfig struct {
	StoreDir string
	HttpPort int
	GenBlock bool
	LogPath  string
}

type TxPoolConfig struct {
	MaxPending   uint32
	MaxInPending uint32
	MaxInQueue   uint32
	MaxTxInPool  uint32
}
type CliqueConfig struct {
	Period uint64
	Epoch  uint64
}

func DecodeConfig(cfg Provider) (c MagneticConfig, err error) {
	m := cfg.GetStringMap("config")
	err = mapstructure.WeakDecode(m, &c)
	return
}

type Provider interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	Get(key string) interface{}
	Set(key string, value interface{})
	IsSet(key string) bool
	WatchConfig()
	OnConfigChange(run func(in fsnotify.Event))
	Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error
}

// FromConfigString creates a config from the given YAML, JSON or TOML config. This is useful in tests.
func FromConfigString(path, configType string) (Provider, error) {
	v := viper.New()
	v.SetConfigType(configType)
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// GetStringSlicePreserveString returns a string slice from the given config and key.
// It differs from the GetStringSlice method in that if the config value is a string,
// we do not attempt to split it into fields.
func GetStringSlicePreserveString(cfg Provider, key string) []string {
	sd := cfg.Get(key)
	if sds, ok := sd.(string); ok {
		return []string{sds}
	} else {
		return cast.ToStringSlice(sd)
	}
}


type Mongo struct {
	Addr      string
	Timeout   time.Duration
	PoolLimit int
	Database  string
}