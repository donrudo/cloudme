package config

// config package for the config file used by the cloudme tool

type App struct {
	Name      string
	Command   string
	Ports     []uint64
	Version   string
	Source    string
	BaseImage string
}

type Mservices struct {
	Name       string
	Image      string
	DockerFile string
	Hostname   string
	Mount      []string
	Command    string
	ConfigPath string
	Requires   string
	Logs       string
}

type Root struct {
	Application   App
	Microservices []Mservices
}
