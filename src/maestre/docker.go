package maestre

// maestre module for the Docker Runtime actions

import ( // Standard deps
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

import ( // internal or vendor deps
	"config"
	DockerAPI "github.com/samalba/dockerclient"
	dbg "github.com/tj/go-debug"
)

type RoutineResult struct {
	Results map[string]chan bool
}

type DockerRuntime struct {
	Runtime Runtime
	Context string
	Api     *DockerAPI.DockerClient
	Debug   dbg.DebugFunction

	// Builded instead of Built to be less confusing with "Build" function
	Builded  RoutineResult
	Started  RoutineResult
	Verified RoutineResult
}

func NewDockerClient(context string) (*DockerRuntime, error) {
	Docker := new(DockerRuntime)
	Docker.Context = context
	Docker.Debug = dbg.Debug("Docker Runtime")
	Docker.Debug("Configuring Docker Runtime")
	tlsConfig := &tls.Config{}
	Docker.Builded.Results = make(map[string]chan bool)
	Docker.Started.Results = make(map[string]chan bool)
	Docker.Verified.Results = make(map[string]chan bool)
	var err error

	if os.Getenv("DOCKER_TLS_VERIFY") != "" && os.Getenv("DOCKER_TLS_VERIFY") != "0" {
		caFile := os.Getenv("DOCKER_CERT_PATH") + "/ca.pem"
		certFile := os.Getenv("DOCKER_CERT_PATH") + "/cert.pem"
		keyFile := os.Getenv("DOCKER_CERT_PATH") + "/key.pem"

		cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
		pemCerts, _ := ioutil.ReadFile(caFile)

		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

		tlsConfig.Certificates = []tls.Certificate{cert}

		tlsConfig.RootCAs.AppendCertsFromPEM(pemCerts)
	} else {
		tlsConfig = nil
	}

	if os.Getenv("DOCKER_HOST") != "" {
		Docker.Api, err = DockerAPI.NewDockerClient(os.Getenv("DOCKER_HOST"), tlsConfig)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		Docker.Debug("Ready Docker Runtime")
	} else {
		err := fmt.Errorf("DOCKER_HOST empty or not set")
		Docker.Debug("FATAL: %s", err)
		return nil, err
	}
	return Docker, nil
}

// Run a container with the given configuration as soon as it gets a true value from the Builded channel.
func (dr DockerRuntime) Run(service config.Mservices, app config.App) {
	// Waits until the build process is began
	// TODO: add timeout, maybe the configuration did not required to re-build the images.
	dr.Started.Results[service.Name] = make(chan bool)
	defer close(dr.Started.Results[service.Name])
	for dr.Builded.Results[service.Image] == nil {
		time.Sleep(500 * time.Millisecond)
	}

	dr.Debug("Wait for Build to Run container: %s", service.Name)
	dbgmsg := fmt.Sprintf("RUN - waiting for Build(%s) to finish", service.Image)
	BuildOk := dr.Builded.WaitForResult(service.Image, dr.Debug, dbgmsg)

	if !BuildOk {
		dr.Debug("RUN - Build(%s) failed", service.Image)
		dr.Started.Results[service.Name] <- false
		return
	}

	// TODO: implement RUN process:
	// -- 1. Get script to execute (clone or pull if is a github repo)
	// -- 2. Delete old container version (causes downtime)
	// -- 3. Run the new container version
	dr.Debug("RUN - Build(%s) succeed, Starting the container %s now!", service.Image, service.Name)
	dr.Started.Results[service.Name] <- true
	return
}

// Verify uses the healthcheck set at the given service to put a message into the channel once it's correctly running or if an error is detected (timeouts at 30 secs)
func (dr DockerRuntime) Verify(service config.Mservices, app config.App) {
	dr.Verified.Results[service.Name] = make(chan bool)
	defer close(dr.Verified.Results[service.Name])
	dr.Debug("Waiting to get a healthy answer from container: %s", service.Name)
	dbgmsg := fmt.Sprintf("VERIFY - waiting for Run(%s) to finish", service.Name)
	StartedOk := dr.Started.WaitForResult(service.Name, dr.Debug, dbgmsg)

	if !StartedOk {
		dr.Debug("Verify - Run(%s) failed", service.Image)
		dr.Verified.Results[service.Name] <- false
		return
	}

	// TODO: implement Verify process HERE.
	dr.Debug("Verify - Run(%s) Succeed, Verifying healthcheck now!", service.Name)
	dr.Verified.Results[service.Name] <- true
	return
}

// Build the docker images to be used, when finish puts a bool value into the Results channel
func (dr DockerRuntime) Build(service config.Mservices, app config.App) {
	dr.Builded.Results[service.Image] = make(chan bool)
	defer close(dr.Builded.Results[service.Image])

	// dockerclient lib relies on a tar file to be used as context, we only put docker files and configurations there.
	dockerBuildContextName, err := dr.CreateTar(dr.Context)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
		dr.Builded.Results[service.Image] <- false
		return
	}
	dockerBuildContext, err := os.Open(dockerBuildContextName)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
		dr.Builded.Results[service.Image] <- false
		return
	}
	// Finishes tar context creation

	// TODO: Implement AuthConfig for docker registry repos
	config := &DockerAPI.ConfigFile{
		Configs: map[string]DockerAPI.AuthConfig{
			"": DockerAPI.AuthConfig{}},
	}

	image := &DockerAPI.BuildImage{
		Config:         config,
		SuppressOutput: false,
		Remove:         true,
		DockerfileName: service.Dockerfile,
		Context:        dockerBuildContext,
		RepoName:       service.Image,
		CgroupParent:   app.Name,
	}

	// Builds the Image with the given configuration
	dr.Debug("Building Image: %s ", service.Name)
	reader, err := dr.Api.BuildImage(image)
	defer dockerBuildContext.Close()
	if err != nil {
		dr.Debug("%s", err)
	} else {
		// When correctly built, we are showing the output at debug, then put a message at the channel
		if b, err := ioutil.ReadAll(reader); err == nil {
			dr.Debug("Build output: %s", string(b))
			dr.Builded.Results[service.Image] <- true
			return
		} else {
			fmt.Println(err)
			dr.Builded.Results[service.Image] <- false
			return
		}
	}
}

// Delete the container related to the given service,
func (dr DockerRuntime) Delete(service config.Mservices, app config.App) {
}

//WaitForResult waits until a message is received from the given channel and closes it.
func (this RoutineResult) WaitForResult(channel string, debugfunc dbg.DebugFunction, debugmsg string) bool {
	done := false
	result := false
	for !done {
		//wait until a message is received from the Results[channel]
		select {
		case result = <-this.Results[channel]:
			done = true
		default:
			debugfunc(debugmsg)
			debugfunc("  using %s", channel)
			time.Sleep(500 * time.Millisecond)
			done = false
		}
	}
	return result
}
