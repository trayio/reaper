package config

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Count: how many instances
// Age: minium age of instance
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

	return generate(data)
}

func generate(data []byte) (Configuration, error) {
	c := make(Configuration)

	err := hcl.Decode(&c, string(data))

	//err := json.Unmarshal(data, &c)
	return c, err
}
