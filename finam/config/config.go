package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Redis        RedisConfig        `yaml:"redis"`
	Notification NotificationConfig `yaml:"notification"`
	Chart        ChartConfig        `yaml:"chart"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
	DB   int    `yaml:"db"`
}

type NotificationConfig struct {
	StartHour   int   `yaml:"start_hour"`
	StartMinute int   `yaml:"start_minute"`
	NSInterval  int64 `yaml:"interval"`
}

type ChartConfig struct {
	NLatest int `yaml:"n_latest"`
}

type Flags struct {
	ConfigPath string
}

func ParseFlags() Flags {
	processorCfgPath := flag.String("config", "", "Path to service cfg")
	flag.Parse()
	return Flags{
		ConfigPath: *processorCfgPath,
	}
}

func MustLoad(cfgPath string, cfg any) {
	if cfgPath == "" {
		log.Fatal("Config path is not set")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist by this path: %s", cfgPath)
	}

	if err := cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		log.Fatalf("error reading config: %s", err)
	}
}
