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

type DockerRuntime struct {
	Runtime Runtime
	Context string
	Api     *DockerAPI.DockerClient
	Debug   dbg.DebugFunction

	// Builded instead of Built to be less confusing with "Build" function
	Builded  map[string]chan bool
	Started  map[string]chan bool
	Verified map[string]chan bool
}

func NewDockerClient(context string) *DockerRuntime {
	Docker := new(DockerRuntime)
	Docker.Context = context
	Docker.Debug = dbg.Debug("Docker Runtime")
	Docker.Debug("Configuring Docker Runtime")
	tlsConfig := &tls.Config{}
	Docker.Builded = make(map[string]chan bool)
	Docker.Started = make(map[string]chan bool)
	Docker.Verified = make(map[string]chan bool)
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
		Docker.Debug("Ready Docker Runtime")
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}

	return Docker
}

func (dr DockerRuntime) Run(service config.Mservices, app config.App) {
	dr.Started[service.Name] = make(chan bool)
	for dr.Builded[service.Image] == nil {
		time.Sleep(500 * time.Millisecond)
	}
	dr.Debug("Wait for Build to Run container: %s", service.Name)
	done := false
	BuildOk := false
	for !done {
		//wait until a message is received from the Builded[service.Image] channel
		select {
		case BuildOk = <-dr.Builded[service.Image]:
			done = true
		default:
			time.Sleep(500 * time.Millisecond)
			dr.Debug("RUN - waiting for Build(%s) to finish", service.Image)
			done = false
		}
	}

	if !BuildOk {
		dr.Debug("RUN - Build(%s) failed", service.Image)
		dr.Started[service.Name] <- false
		return
	}
}

// Verify uses the healthcheck set at the given service to put a message into the channel once it's correctly running or if an error is detected (timeouts at 30 secs)
func (dr DockerRuntime) Verify(service config.Mservices, app config.App) {
	dr.Debug("Waiting to get a healthy answer from container: %s", service.Name)
}

func (dr DockerRuntime) Build(service config.Mservices, app config.App) {
	dr.Builded[service.Image] = make(chan bool)
	dockerBuildContextName, err := dr.CreateTar(dr.Context)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
		dr.Builded[service.Image] <- false
		return
	}
	dockerBuildContext, err := os.Open(dockerBuildContextName)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
		dr.Builded[service.Image] <- false
		return
	}

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

	dr.Debug("Building Image: %s ", service.Name)
	reader, err := dr.Api.BuildImage(image)
	defer dockerBuildContext.Close()
	if err != nil {
		dr.Debug("%s", err)
	} else {
		if b, err := ioutil.ReadAll(reader); err == nil {
			dr.Debug("Build output: %s", string(b))
			dr.Builded[service.Image] <- true
			return
		} else {
			fmt.Println(err)
			dr.Builded[service.Image] <- false
			return
		}
	}
}

func (dr DockerRuntime) Delete(service config.Mservices, app config.App) {
}
