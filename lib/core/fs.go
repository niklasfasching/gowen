package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func spit(path string, content string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	assert(err == nil, "error creating directory for path %s", path)
	err = ioutil.WriteFile(path, []byte(content), os.ModePerm)
	assert(err == nil, "error creating file for path %s", path)
}

func slurp(path string) string {
	bs, err := ioutil.ReadFile(path)
	assert(err == nil, "error reading file %s", path)
	return string(bs)
}
