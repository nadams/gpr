package gps

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_Brightness(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "test1.gps"))
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	c, err := Brightness(f)
	if err != nil {
		t.Error(err)
	}

	if c != 76 {
		t.Error("brightness has incorrect value")
	}
}

func Test_Contrast(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "test1.gps"))
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	c, err := Contrast(f)
	if err != nil {
		t.Error(err)
	}

	if c != 87 {
		t.Error("contrast has incorrect value")
	}
}
