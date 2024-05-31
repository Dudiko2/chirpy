package db

import (
	"fmt"
	"os"
	"testing"
)

func failOnErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewDB(t *testing.T) {
	fName := "db_test.json"
	d, err := NewDB(fName)
	failOnErr(t, err)
	t.Logf("File %s created", fName)
	defer func() {
		err := os.Remove(fName)
		failOnErr(t, err)
		t.Logf("Removed %s", fName)
	}()
	if d.path != fName {
		failOnErr(t, fmt.Errorf("path should be %s", fName))
	}
	c, err := d.CreateChirp("hello there")
	if err != nil {
		failOnErr(t, err)
	}
	t.Logf("Created chirp %v", c)
}
