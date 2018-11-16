//+build ignore

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/niklasfasching/gowen/lib/core"
)

var packages = map[string]string{
	"strings": "strings",
	"strconv": "strconv",
	"ioutil":  "io/ioutil",
	"time":    "time",
	"os":      "os",
	"exec":    "os/exec",
}

func main() {
	err := ioutil.WriteFile("go_packages__generated__.go",
		[]byte(core.GenerateGoPackageRegisterFileContent("core", packages)),
		os.ModePerm,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("inlined_gowen__generated__.go",
		[]byte(core.GenerateGowenInlineFileContent("core", []string{"core.gow"})),
		os.ModePerm,
	)
	if err != nil {
		log.Fatal(err)
	}
}
