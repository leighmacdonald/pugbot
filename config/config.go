package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

const (
	CfgLogLevel = "log.level"
	CfgAuthId   = "auth.id"
	//CfgAuthSecret = "auth.secret"
	CfgAuthPerms = "auth.perms"
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

func OAuthUrl() string {
	cfg := GetConfig()
	return fmt.Sprintf(
		"https://discordapp.com/api/oauth2/authorize?client_id=%d&scope=bot&permissions=%d",
		cfg.GetInt64(CfgAuthId),
		cfg.GetInt64(CfgAuthPerms),
	)
}

func GetConfig() *viper.Viper {
	if !configured {
		Init("config")
	}
	return config
}
