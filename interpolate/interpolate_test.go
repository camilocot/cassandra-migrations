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

	if err != nil {
		log.Fatal(err)
	}

	fn := dir + "/testfile.cql"

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

func TestInterpolate(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)

	}
	fn1 := dir + "/testfile1.cql"
	fn2 := dir + "/testfile2.cql"

	defer os.RemoveAll(dir)

	_, _ = os.Create(fn1)
	_, _ = os.Create(fn2)

	called := 0
	var cFiles []string

	Interpolate(dir, "old", "new", func(f, old, new string) error {
		if old != "old" {
			t.Errorf("Unexpected parameter in old %s", old)
		}
		if new != "new" {
			t.Errorf("Unexpected parameter in new %s", new)
		}
		cFiles = append(cFiles, f)
		called++

		return nil
	})

	if !contains(cFiles, fn1) {
		t.Errorf("File %s was not parsed", fn1)
	}

	if !contains(cFiles, fn2) {
		t.Errorf("File %s was not parsed", fn2)
	}

	if called != 2 {
		t.Errorf("Invalid number of parsed files")
	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
