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
	//authConfig DockerAPI.AuthConfig
	Api   *DockerAPI.DockerClient
	Debug dbg.DebugFunction
}

func NewDockerClient(context string) *DockerRuntime {
	Docker := new(DockerRuntime)
	Docker.Context = context
	Docker.Debug = dbg.Debug("Docker Runtime")
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

	return Docker
}

func (dr DockerRuntime) Run(service config.Mservices, app config.App) {

}

// Verify uses the healthcheck set at the given service to put a message into the channel once it's correctly running or if an error is detected (timeouts at 30 secs)
func (dr DockerRuntime) Verify(service config.Mservices, app config.App) {
}

func (dr DockerRuntime) Build(service config.Mservices, app config.App) {
	var err error

	dockerBuildContextName, err := dr.CreateTar(dr.Context)
	if err != nil {
		dr.Debug("%s", err)
		Error <- err
	}
	dockerBuildContext, err := os.Open(dockerBuildContextName)
	if err != nil {
		dr.Debug("%s", err)
		Error <- err
	}

	dr.Debug("Building Images, base Image: %s ", app.BaseImage)

	if err != nil {
		dr.Debug("%s", err)
		Error <- err
	}
	defer dockerBuildContext.Close()

	image := &DockerAPI.BuildImage{
		SuppressOutput: true,
		Remove:         true,
		DockerfileName: service.Dockerfile,
		Context:        dockerBuildContext,
		RepoName:       service.Image,
		CgroupParent:   app.Name,
	}

	reader, err := dr.Api.BuildImage(image)
	defer dockerBuildContext.Close()
	if err != nil {
		dr.Debug("%s", err)
		Error <- err
	} else {
		if b, err := ioutil.ReadAll(reader); err == nil {
			dr.Debug("Build output: %s", string(b))
		} else {
			fmt.Println(err)
			Error <- err
		}
	}
}

func (dr DockerRuntime) Delete(service config.Mservices, app config.App) {
}
