package maestre

// maestre manages the container engines, currently only docker but rkt is planned too

import (
	"config"
	"encoding/json"
	. "github.com/tj/go-debug"
	"os"
	"path/filepath"
	"time"
)

type Runtime interface {
	Run(config.Mservices, config.App)
	Verify(config.Mservices, config.App)
	Build(config.Mservices, config.App)
	Delete(config.Mservices, config.App)
}

var (
	configPath string
	Error      chan error
	Output     chan error
	debug      = Debug("Runtime Maestre")
	Rt         Runtime
	driver     = "docker"
	Config     *config.Root
)

func Init(runtime string, configFile string) error {
	Error = make(chan error)
	Output = make(chan error)
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

	debug("loaded configuration file")
	return cfg, nil
}

func Build() (int, error) {
	var i int
	for i = 0; i < len(Config.Microservices); i++ {
		go Rt.Build(Config.Microservices[i], Config.Application)
	}
	return 0, nil
}

func Deploy() (int, error) {
	var i int
	for i = 0; i < len(Config.Microservices); i++ {
		go Rt.Build(Config.Microservices[i], Config.Application)
		go Rt.Run(Config.Microservices[i], Config.Application)
		go Rt.Verify(Config.Microservices[i], Config.Application)
	}

	var isDone bool
	for isDone = false; isDone == false; isDone = false {
		select {
		case errs := <-Error:
			isDone = true
			debug("%s", errs)
		default:
			time.Sleep(time.Second * 1)
		}
	}
	close(Error)
	return 0, nil
}
