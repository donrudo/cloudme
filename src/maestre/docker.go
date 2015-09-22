package maestre

// maestre module for the Docker Runtime actions

import ( // Standard deps

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
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
}

func NewDockerClient(context string) *DockerRuntime {
	Docker := new(DockerRuntime)
	Docker.Context = context
	Docker.Debug = dbg.Debug("Docker Runtime")
	Docker.Debug("Configuring Docker Runtime")
	tlsConfig := &tls.Config{}
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

	Docker.Api, err = DockerAPI.NewDockerClient(os.Getenv("DOCKER_HOST"), tlsConfig)
	if err != nil {
		fmt.Println(err)
	}

	Docker.Debug("Ready Docker Runtime")
	return Docker
}

func (dr DockerRuntime) Run(service config.Mservices, app config.App) {
	dr.Debug("Wait for Build to Run container: %s", service.Name)
}

// Verify uses the healthcheck set at the given service to put a message into the channel once it's correctly running or if an error is detected (timeouts at 30 secs)
func (dr DockerRuntime) Verify(service config.Mservices, app config.App) {
	dr.Debug("Waiting to get a healthy answer from container: %s", service.Name)
}

func (dr DockerRuntime) Build(service config.Mservices, app config.App) {
	dockerBuildContextName, err := dr.CreateTar(dr.Context)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
	}
	dockerBuildContext, err := os.Open(dockerBuildContextName)
	if err != nil {
		dr.Debug("%s", err)
		fmt.Println(err)
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
		} else {
			fmt.Println(err)
		}
	}
}

func (dr DockerRuntime) Delete(service config.Mservices, app config.App) {
}
