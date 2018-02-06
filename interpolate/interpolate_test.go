package interpolate

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestWalk(t *testing.T) {
	content := []byte("temp file")
	dir, err := ioutil.TempDir("", "test")
	fn := "tempfile"
	expectedFn := dir + "/" + fn

	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, fn)
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		log.Fatal(err)
	}

	fileList, err := Walk(dir)

	if err != nil {
		log.Fatal(err)
	}

	if len(fileList) != 1 {
		t.Error("FileList has an incorrect number of elements")
	}

	if fileList[0] != expectedFn {
		t.Errorf("FileList element name is not right: %s vs %s", fileList[0], expectedFn)
	}

}

func TestReplaceNoCQL(t *testing.T) {
	content := []byte("temp ${file}")
	dir, err := ioutil.TempDir("", "test")
	fn := dir + "/testfile"

	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := ioutil.WriteFile(fn, content, 0666); err != nil {
		log.Fatal(err)
	}

	err = Replace(fn, "file", "new")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	c, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Error(err)
	}

	if string(content) != string(c) {
		t.Errorf("Unexpected content in %s, content: %s", fn, c)
	}
}

func TestReplaceCQL(t *testing.T) {
	content := []byte("temp ${file}")
	dir, err := ioutil.TempDir("", "test")
	fn := dir + "/testfile.cql"

	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := ioutil.WriteFile(fn, content, 0666); err != nil {
		log.Fatal(err)
	}

	err = Replace(fn, "file", "new")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	c, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Error(err)
	}

	if string(content) != string(c) {
		t.Errorf("Unexpected content in %s, content: %s", fn, c)
	}

}
