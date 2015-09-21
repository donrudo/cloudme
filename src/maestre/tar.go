package maestre

// source based on https://www.socketloop.com/tutorials/golang-archive-directory-with-tar-and-gzip

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (dr DockerRuntime) CreateTarCMD(sourcedir string) (*os.File, error) {

	cmd := exec.Command("tar", "-cf")
	cmd.Stdin = strings.NewReader("/tmp/" + sourcedir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return os.Open(sourcedir)
}

func (dr DockerRuntime) CreateTar(sourcedir string) (string, error) {

	os.Chdir(sourcedir)
	dir, err := os.Open(".")
	if err != nil {
		return "", err
	}
	defer dir.Close()

	files, err := dir.Readdir(0) // grab the files list
	if err != nil {
		return "", err
	}

	tarfile, err := ioutil.TempFile("/tmp", "docker")
	if err != nil {
		dr.Debug("%s", err)
	}
	tarName := tarfile.Name()
	var fileWriter io.WriteCloser = tarfile

	tarfileWriter := tar.NewWriter(fileWriter)
	defer tarfileWriter.Close()

	for _, fileInfo := range files {

		if fileInfo.IsDir() {
			continue
		}

		// see https://www.socketloop.com/tutorials/go-file-path-independent-of-operating-system

		file, err := os.Open(dir.Name() + string(filepath.Separator) + fileInfo.Name())
		if err != nil {
			return "", err
		}

		defer file.Close()

		// prepare the tar header

		header := new(tar.Header)
		header.Name = file.Name()
		header.Size = fileInfo.Size()
		header.Mode = int64(fileInfo.Mode())
		header.ModTime = fileInfo.ModTime()

		err = tarfileWriter.WriteHeader(header)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(tarfileWriter, file)
		if err != nil {
			return "", err
		}

	}

	return tarName, nil
}
