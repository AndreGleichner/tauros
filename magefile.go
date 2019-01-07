// +build mage
// https://magefile.org/
// Sample files:
// https://github.com/gohugoio/hugo/blob/master/magefile.go
// https://github.com/wrouesnel/postgres_exporter/blob/master/magefile.go

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/magefile/mage/mg"     // https://godoc.org/github.com/magefile/mage/mg
	"github.com/magefile/mage/sh"     // https://godoc.org/github.com/magefile/mage/sh
	"github.com/magefile/mage/target" // https://godoc.org/github.com/magefile/mage/target
)

var toolsEnvWindows = map[string]string{"GOOS": "windows", "GOARCH": "386"}

//var toolsEnvWindows = map[string]string{"GOOS": "windows", "GOARCH": "amd64"}
var toolsEnvLinux = map[string]string{"GOOS": "linux", "GOARCH": "amd64"}

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

func buildCmd(cmd string, linux bool) error {
	fmt.Println("Running go build in dir ./cmd/" + cmd)

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)

	err := os.Chdir("./cmd/" + cmd)

	if linux {
		if err = sh.RunWith(toolsEnvLinux, "go", "build"); err != nil {
			return err
		}

		return os.Rename(cmd, pwd+"/bin/"+cmd)
	} else {
		if err = sh.RunWith(toolsEnvWindows, "go", "build"); err != nil {
			return err
		}

		return os.Rename(cmd+".exe", pwd+"/bin/"+cmd+".exe")
	}
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	pwd, _ := os.Getwd()

	mg.Deps(InstallDeps)
	fmt.Println("Building... in dir " + pwd)

	var err error

	changed, err := target.Path("api/api.pb.go", "api/api.proto")
	if changed && (err == nil) {
		fmt.Println("Compiling api.proto")

		if err = sh.Run("protoc", "--go_out=plugins=grpc:.", "api/api.proto"); err != nil {
			return err
		}
	}

	if err = buildCmd("tauros_proxy", false); err != nil {
		return err
	}
	if err = buildCmd("tauros_vm", false); err != nil {
		return err
	}
	if err = buildCmd("tauros_vm", true); err != nil {
		return err
	}

	return err
}

func InstallDeps() error {
	fmt.Println("Installing Deps...")

	protocDst := exePath(os.Getenv("GOPATH")+"/bin", "protoc")
	if _, err := os.Stat(protocDst); os.IsNotExist(err) {
		fmt.Println("protoc tool not found")

		// Install protoc to GOPATH bin
		var protocUrl string
		if isWindows() {
			protocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-win32.zip"
		} else {
			protocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip"
		}

		tempDir, err := tempDir()
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		tempProtocZipFile := filepath.Join(tempDir, "protoc.zip")

		err = downloadFile(tempProtocZipFile, protocUrl)
		if err != nil {
			return err
		}

		_, err = unzipAll(tempProtocZipFile, tempDir)
		if err != nil {
			return err
		}

		src := exePath(tempDir+"/bin", "protoc")

		err = os.Rename(src, protocDst)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("protoc tool already installed here: " + protocDst)
	}

	return nil
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")

}

func createDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func tempDir() (string, error) {
	dir := path.Join(os.TempDir(), uuid.New().String())
	err := createDir(dir)
	return dir, err
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}
func exePath(dir string, name string) string {
	path := filepath.Join(dir, name)
	if isWindows() {
		path += ".exe"
	}
	return path
}

func downloadFile(filepath string, url string) error {
	fmt.Println("Downloading from " + url)

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func unzipAll(zipFile string, destDir string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(destDir, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}
