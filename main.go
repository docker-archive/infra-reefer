package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
)

const templateSuffix = ".tmpl"

type list []string

func (l *list) String() string {
	return ""
}

func (l *list) Set(str string) error {
	*l = append(*l, str)
	return nil
}

type templateData struct {
}

func (td templateData) Env(str string) string {
	return os.Getenv(str)
}

type templateList map[string]target

type target struct {
	template *template.Template
	info     os.FileInfo
}

func (tl templateList) String() string {
	return ""
}

func (tl templateList) Set(str string) error {
	parts := strings.SplitN(str, ":", 2)
	stat, err := os.Stat(parts[0])
	if err != nil {
		return err
	}
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
	tl[dest] = target{
		template: t,
		info:     stat,
	}
	return nil
}

func (tl templateList) Render(root string) error {
	for d, t := range tl {
		dest := d
		if !path.IsAbs(d) {
			dest = path.Join(root, d)
		}
		data := templateData{}
		if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return fmt.Errorf("Couldn't mkdir %s: %s", filepath.Dir(dest), err)
		}
		fh, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("Couldn't create %s: %s", dest, err)
		}
		if err := fh.Chmod(t.info.Mode()); err != nil {
			return err
		}
		if err := t.template.Execute(fh, data); err != nil {
			return err
		}
	}
	return nil
}

func getFilteredEnv(keep []string) (env []string) {
	for _, k := range keep {
		v := os.Getenv(k)
		if v == "" {
			continue
		}
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

var (
	execEnv = []string{}
	keepAllEnvs = flag.Bool("E", false, "Keep all environment variables")
	keepEnvs = list{
		"COLORS",
		"DISPLAY",
		"HOME",
		"HOSTNAME",
		"KRB5CCNAME",
		"LS_COLORS",
		"PATH",
		"PS1",
		"PS2",
		"TZ",
		"XAUTHORITY",
		"XAUTHORIZATION",
	}
	root = flag.String("r", "", "Root for relative paths")
)

func main() {
	templates := templateList{}
	flag.Var(&templates, "t", "Specify template and append optional destination after collons. Format: foo.tmpl:/etc/foo.conf")
	flag.Var(&keepEnvs, "e", fmt.Sprintf("Keep specified environment variables beside %s", strings.Join(keepEnvs, ",")))
	flag.Parse()
	if *root == "" {
		r, err := os.Getwd()
		if err != nil {
			log.Fatal("Not root (-r) specified and couldn't get working directory")
		}
		*root = r
	}
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("No command provided, exiting")
	}

	templates.Render(*root)
	path, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatal(err)
	}
	if *keepAllEnvs {
		execEnv = os.Environ()
	} else {
		execEnv = getFilteredEnv(keepEnvs)
	}
	if err := syscall.Exec(path, args, execEnv); err != nil {
		log.Fatal(err)
	}
}
