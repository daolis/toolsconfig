package toolsconfig

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var readConfiguration = defaultReadConfiguration

func defaultReadConfiguration() *Config {
	var config Config
	filename := viper.ConfigFileUsed()
	file, err := os.Open(filename)
	if err != nil {
		return &Config{}
	}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal(err)
	}
	return &config
}

var saveConfiguration = defaultSaveConfiguration

func defaultSaveConfiguration(config *Config) error {
	dir := path.Dir(viper.ConfigFileUsed())
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0700)
	}

	file, err := os.Create(viper.ConfigFileUsed())
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(4)
	if err := encoder.Encode(config); err != nil {
		log.Fatal(err)
	}
	return nil
}

// by default it's ~/.toolsconfig/config.yaml
var configFile = defaultConfigFileName

func defaultConfigFileName(dir, file string) (*string, error) {
	var filePath string
	if path.IsAbs(dir) {
		filePath = path.Join(dir, file)
		return &filePath, nil
	}
	if dir == "." {
		fileName := path.Join(dir, file)
		return &fileName, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fileName := path.Join(home, dir, file)
	return &fileName, nil
}

func toEnvironmentKey(key string, additionalElements ...string) string {
	prefix := strings.ReplaceAll(key, ".", "_")
	prefix = strings.ReplaceAll(prefix, "-", "_")
	return strings.ToUpper(strings.Join(append([]string{prefix}, additionalElements...), "_"))
}
