package maestre

// maestre manages the container engines, currently only docker but rkt is planned too

import (
	"config"
	"encoding/json"
	. "github.com/tj/go-debug"
	"os"
	"path/filepath"
)

type Runtime interface {
	Run(config.Root) error
	Create(config.Root) error
	Build(config.Root) error
	Delete(config.Root) error
}

var (
	configPath string
	debug      = Debug("Runtime Maestre")
	Rt         Runtime
	driver     = "docker"
	Config     *config.Root
)

func Init(runtime string, configFile string) error {
	cfg, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	Config = cfg

	switch driver {
	case "docker":
		debug("Docker Runtime Enabled")
		Rt = NewDockerClient(configPath)
	case "rkt":
		//not yet supported
		// runtime = new(RktRuntime)
	}

	return nil
}

func loadConfig(Path string) (*config.Root, error) {
	file, err := os.Open(Path)
	defer file.Close()
	configPath = filepath.Dir(Path)
	if err != nil {
		return &config.Root{}, err
	}

	cfg := &config.Root{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func Build() (uint32, error) {
	err := Rt.Build(*Config)
	if err != nil {
		return 1, err
	}

	return 0, nil
}
