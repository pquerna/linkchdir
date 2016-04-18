package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const testTemplate = `
package {{.Name}}

import (
	"testing"
	"time"
)

func TestFoo(t *testing.T) {
	time.Sleep(time.Microsecond)
	t.Skip()
}
`

const pkgPrefix = "linkchdir"
const ntests = 500
const nruns = 100

type ctx struct {
	Name string
}

var t = template.Must(template.New("test.go").Parse(testTemplate))

func buildtestdir(tdir string) []string {
	rv := []string{}
	for i := 0; i < ntests; i++ {
		pkgname := fmt.Sprintf("p%d", i)
		pdir := filepath.Join(tdir, pkgname)
		os.MkdirAll(pdir, 0755)
		buf := &bytes.Buffer{}

		err := t.Execute(buf, &ctx{
			Name: pkgname,
		})

		if err != nil {
			panic(err)
		}
		ioutil.WriteFile(filepath.Join(pdir, "p_test.go"), buf.Bytes(), 0644)
		rv = append(rv, pkgPrefix+"/"+pkgname)
	}

	return rv
}

func main() {
	var err error
	var out []byte
	var d string

	defer func() {
		if err == nil {
			os.RemoveAll(d)
		}
	}()

	d, err = ioutil.TempDir("", "linkchdir-test")
	if err != nil {
		panic(err)
	}

	tdir := filepath.Join(d, "src", pkgPrefix)

	println("tests directory: " + tdir)

	tests := buildtestdir(tdir)

	cmd := []string{"test", "-v", "-p", "20"}
	cmd = append(cmd, tests...)

	os.Chdir(tdir)

	os.Setenv("GOPATH", d+string(filepath.ListSeparator)+os.Getenv("GOPATH"))
	for i := 0; i < nruns; i++ {
		c := exec.Command("go", cmd...)
		fmt.Printf("running %d\n", i)
		out, err = c.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed at test %d\n", i)
			println(string(out))
			panic(err)
		}
	}
}
