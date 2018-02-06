package interpolate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type convert func(string, os.FileInfo, string, string) error

func Walk(searchDir string) ([]string, error) {

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if fs, err := os.Stat(path); err == nil && !fs.IsDir() {
			fileList = append(fileList, path)
		}
		return err
	})

	if e != nil {
		panic(e)
	}

	return fileList, nil
}

func Replace(fi string, old string, new string) (err error) {

	fstat, err := os.Stat(fi)
	if err != nil {
		panic(err)

	}
	if !!fstat.IsDir() {
		return nil
	}

	matched, err := filepath.Match("*.cql", fi)

	if err != nil {
		panic(err)
	}

	if matched {
		read, err := ioutil.ReadFile(fi)
		if err != nil {
			panic(err)
		}

		newContents := strings.Replace(string(read), "${"+old+"}", new, -1)

		err = ioutil.WriteFile(fi, []byte(newContents), 0)
		if err != nil {
			panic(err)
		}

	}

	return nil
}

func Interpolate(searchDir string, old string, new string, fn convert) error {
	fileList, err := Walk(searchDir)
	if err != nil {
		return err
	}

	for _, f := range fileList {
		fInfo, err := os.Stat(f)
		if err != nil {
			return err
		}
		err = fn(searchDir, fInfo, old, new)
		if err != nil {
			return err
		}
	}
	return nil
}
