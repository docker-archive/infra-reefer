package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
)

const templateSuffix = ".tmpl"

type templateData struct {
}

func (td templateData) Env(str string) string {
	return os.Getenv(str)
}

type templateList map[string]*template.Template

func (tl templateList) String() string { return "" }

func (tl templateList) Set(str string) error {
	parts := strings.SplitN(str, ":", 2)
	t, err := template.ParseFiles(parts[0])
	if err != nil {
		return err
	}
	dest := ""
	if len(parts) == 2 {
		dest = parts[1]
	} else {
		dest = strings.TrimSuffix(parts[0], templateSuffix)
	}
	tl[dest] = t
	return nil
}

func (tl templateList) Render(root string) error {
	for d, t := range tl {
		dest := path.Join(root, d)
		data := templateData{}
		if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return fmt.Errorf("Couldn't mkdir %s: %s", filepath.Dir(dest), err)
		}
		fh, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("Couldn't create %s: %s", dest, err)
		}
		if err := t.Execute(fh, data); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	templates := templateList{}
	flag.Var(&templates, "t", "Specify template and append optional destination after collons. Format: foo.tmpl:/etc/foo.conf")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("No command provided, exiting")
	}

	templates.Render("/")
	panic(syscall.Exec(args[0], args, []string{}))
}
