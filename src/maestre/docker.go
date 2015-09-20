package maestre

// maestre module for the Docker Runtime actions

import ( // Standard deps

	"net"
	"net/http"
	"net/url"
	"time"

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

func (dr DockerRuntime) Run(cfg config.Root) error {
	var err error
	return err
}

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}

func (dr DockerRuntime) Create(cfg config.Root) error {
	var err error
	return err
}

func (dr DockerRuntime) GetNewBuildImage(dockerFile string) (*DockerAPI.BuildImage, error) {

	rawimage := &DockerAPI.BuildImage{
		SuppressOutput: true,
		Remove:         true,
		DockerfileName: dockerFile,
	}

	return rawimage, nil
}

func (dr DockerRuntime) Build(cfg config.Root) error {
	var err error

	dr.Debug("Building Images, base Image: %s ", cfg.Application.BaseImage)
	var i int
	for i = 0; i < len(cfg.Microservices); i++ {
		image, err := dr.GetNewBuildImage(cfg.Microservices[i].DockerFile)
		if err != nil {
			fmt.Println(err)
		}

		dockerBuildContext, err := os.Open(dr.Context + ".tar")
		defer dockerBuildContext.Close()
		if err != nil {
			dr.Debug("%s", err)
		}
		image.Context = dockerBuildContext
		image.RepoName = cfg.Microservices[i].Image
		image.CgroupParent = cfg.Application.Name

		reader, err := dr.Api.BuildImage(image)
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
	return err
}

func (dr DockerRuntime) Delete(cfg config.Root) error {
	var err error
	return err
}
