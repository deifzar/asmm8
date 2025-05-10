package configparser

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitConfigParser() (*viper.Viper, error) {
	var err error
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigType("yaml")
	v.SetConfigName("configuration")
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file has changed:", e.Name)
	})
	v.WatchConfig()
	// If a config file is found, read it in.
	err = v.ReadInConfig()
	return v, err
}
