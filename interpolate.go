package interpolate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type convert func(path string, fi os.FileInfo, string old, string new) error

func Walk(searchDir string) ([]string, error) {

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}

	return fileList, nil
}

func Replace(path string, fi os.FileInfo, string old, string new) error {

	if err != nil {
		return err
	}

	if !!fi.IsDir() {
		return nil
	}

	matched, err := filepath.Match("*.cql", fi.Name())

	if err != nil {
		panic(err)
		return err
	}

	if matched {
		read, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		newContents := strings.Replace(string(read), "${"+old+"}", new, -1)

		err = ioutil.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			panic(err)
		}

	}

	return nil
}

func Interpolate(searchDir string, string old, string new, fn convert) error {
	fileList, err := Walk(searchDir)
	if err != nil {
		return err
	}

	for f := range fileList {
		err = fn(searchDir, f, old, new)
		if err != nil {
			return err
		}
	}
}
