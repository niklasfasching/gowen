package core

import (
	"fmt"
	"go/importer"
	"go/types"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/niklasfasching/gowen"
)

var r1 = regexp.MustCompile("(.)([A-Z][a-z]+|[0-9]+)")
var r2 = regexp.MustCompile("([a-z0-9])([A-Z])")
var r3 = regexp.MustCompile("[-]+")

func toLispCase(s string) string {
	s = r1.ReplaceAllString(s, "$1-$2")
	s = r2.ReplaceAllString(s, "$1-$2")
	s = strings.ToLower(s)
	s = strings.Replace(s, "_", "-", -1)
	s = r3.ReplaceAllString(s, "-")
	return s
}

var goPackageRegisterTemplate = `// Code generated automatically via gowen/cmd/generate. DO NOT EDIT.

package %s

import "github.com/niklasfasching/gowen"

%s

func init() {
  gowen.Register(%s, "")
}`

var goInlineGowenTemplate = `// Code generated automatically via gowen/cmd/generate. DO NOT EDIT.

package %s

import "github.com/niklasfasching/gowen"

func init() {
    gowen.Register(nil, %q)
}`

func GenerateGoPackageRegisterFileContent(packageName string, packages map[string]string) string {
	values := "map[string]interface{}{\n"
	imports := "import (\n"
	importer := importer.Default()
	for alias, pkgName := range packages {
		imports += fmt.Sprintf("    %q\n", pkgName)
		pkg, err := importer.Import(pkgName)
		if err != nil {
			log.Fatal(err)
		}
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			object := scope.Lookup(name)
			if _, isType := object.(*types.TypeName); isType || !object.Exported() {
				continue
			}
			key := alias + "/" + toLispCase(name)
			values += fmt.Sprintf("		%q: %s,\n", key, alias+"."+name)
		}
	}
	imports += ")\n"
	values += "    }"
	return fmt.Sprintf(goPackageRegisterTemplate, packageName, imports, values)
}

func GenerateGowenInlineFileContent(packageName string, filenames []string) string {
	input := ""
	for _, f := range filenames {
		bs, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		input += string(bs)
	}
	nodes, err := gowen.Parse(input)
	if err != nil {
		log.Fatal(err)
	}
	output := ""
	for _, n := range nodes {
		output += n.String()
	}
	return fmt.Sprintf(goInlineGowenTemplate, packageName, output)
}
