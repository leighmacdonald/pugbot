package config

import (
	"github.com/spf13/viper"
	"log"
)

const (
	CfgLogLevel = "log.level"
	//CfgAuthId     = "auth.id"
	//CfgAuthSecret = "auth.secret"
	//CfgAuthPerms  = "auth.perms"
	CfgAuthToken = "auth.token"
)

var (
	config     *viper.Viper
	configured = false
)

// Init is an exported method that takes the environment starts the viper
// (external lib) and returns the configuration struct.
func Init(env string) {
	var err error
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(env)
	v.AddConfigPath(".")
	v.AddConfigPath("../config/")
	v.AddConfigPath("config/")
	if err = v.ReadInConfig(); err != nil {
		log.Fatal("error on parsing configuration file")
	}
	config = v

	// Logger
	config.SetDefault(CfgLogLevel, "info")
	configured = true
}

func GetConfig() *viper.Viper {
	if !configured {
		Init("config")
	}
	return config
}
