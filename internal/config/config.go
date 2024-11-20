package config

import (
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Env           string        `yaml:"env"`
	StoragePath   string        `yaml:"storage_path"`
	Redis         Redis         `yaml:"redis"`
	ImageSettings ImageSettings `yaml:"image_settings"`
	HttpServer    HttpServer    `yaml:"http_server"`
	ReqLimit      ReqLimit      `yaml:"req_limit"`
}

type Redis struct {
	Addr     string        `yaml:"addr"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	TTL      time.Duration `yaml:"ttl"`
}

type ImageSettings struct {
	MaxWidth  int    `yaml:"maxWidth"`
	MaxHeight int    `yaml:"maxHeight"`
	Chars     string `yaml:"chars"`
}

type HttpServer struct {
	Addr        string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type ReqLimit struct {
	MaxNumReq uint32        `yaml:"max_num_req"`
	TimeSlice time.Duration `yaml:"time_slice"`
	BanTime   time.Duration `yaml:"ban_time"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
