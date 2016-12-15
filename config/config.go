package config

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Count: how many instances
// Age: minium age of instance
// Region: AWS region
type ConfigItem struct {
	Count  int
	Age    int
	Region string
}

type Configuration map[string]ConfigItem

func New(filename string) (Configuration, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := generate(data)
	if err != nil {
		return cfg, err
	}

	return validate(cfg)
}

func validate(config Configuration) (Configuration, error) {
	for group, cfg := range config {
		if cfg.Count <= 0 {
			return config, fmt.Errorf("%s: count cannot be zero or negative value", group)
		}

		if cfg.Age <= 0 {
			return config, fmt.Errorf("%s: age cannot be zero or negative value", group)
		}

		if cfg.Region == "" {
			return config, fmt.Errorf("%s: invalid region", group)
		}
	}

	return config, nil
}

func generate(data []byte) (Configuration, error) {
	c := make(Configuration)

	err := hcl.Decode(&c, string(data))

	//err := json.Unmarshal(data, &c)
	return c, err
}
