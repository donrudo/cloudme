package main

// Cloudme goal is to ease the deployment of microservices apps using a simple CLI to
// perform the Environment setup (cloudy) and then orchestrating the Containers
// execution (maestre) performing the following procedure:
//	Pre warming steps:
//	 1. Validate config file (maestre)
//   2. Dockerclient library requires that the app folder to be converted to .tar for the building process.

//	Execution Steps:
//	 0.a TODO: Creates the required Instance (cloudy).
//	 0.b TODO: - Setup dependencies (if any)
//
//   1. Build.  (maestre)
//   2. Pull specific commit. (maestre)
//	 3. Kill container (if any). (maestre)
//   4. Run container. (maestre)
//	 5. TODO:  Test healthcheck. (maestre) (only http 200 checks)
//   6. TODO:  If the healthcheck works register container at the given LB. (cloudy)

import ( // standard deps
	"flag"
	"fmt"
	"os"
)

import ( // internal or vendors
	dbg "github.com/tj/go-debug"
	"maestre"
)

var (
	debug      = dbg.Debug("main")
	runtime    = flag.String("runtime", "docker", "Sets the container runtime to be used, defaults to docker")
	command    = flag.String("cmd", "", "Tell cloudme what step to run")
	configFile = flag.String("config", "", "Path to the config file to use")
	secrets    = flag.String("secret", "", "Path to the access config file to use")
)

// Init Validates the basic params cmd is empty then fails then it will initialize the runtime to be used (docker by default)
func Init() {
	flag.Parse()
	debug("Start Params: cmd: %s, config: %s, secrets: %s", *command, *configFile, *secrets)
	if *command == "" {
		os.Exit(1)
	}

	if *configFile == "" {
		os.Exit(1)
	}
	if *runtime == "" {
		os.Exit(1)
	}

	err := maestre.Init(*runtime, *configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Init()

	var exitCode int
	var err error

	switch *command {
	case "deploy":
		fmt.Println("Deploying Application:   " + maestre.Config.Application.Name)
		fmt.Println(" - Application Version: " + maestre.Config.Application.Version)

		exitCode, err = maestre.Deploy()

	case "build":
		fmt.Println("Building Application:   " + maestre.Config.Application.Name)
		fmt.Println(" - Application Version: " + maestre.Config.Application.Version)
		exitCode, err = maestre.Build()

	case "create":
	case "cleanup":
	case "getlogs":
	default:
		fmt.Println("Unsupported command: " + *command)
	}
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}
