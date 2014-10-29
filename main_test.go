package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	expectedContent = "Hello bar\n"
)

func TestExplicitConf(t *testing.T) {
	templates := templateList{}
	os.Setenv("FOO", "bar")
	if err := templates.Set("fixtures/foo.tmpl:/etc/foo.conf"); err != nil {
		t.Fatal(err)
	}
	test(t, templates, "/etc/foo.conf")
}

func TestImplicitConf(t *testing.T) {
	templates := templateList{}
	os.Setenv("FOO", "bar")
	if err := templates.Set("fixtures/foo.tmpl"); err != nil {
		t.Fatal(err)
	}
	test(t, templates, "/fixtures/foo")
}

func test(t *testing.T, templates templateList, dest string) {
	testRoot, err := ioutil.TempDir("/tmp", "test-reefer")
	if err != nil {
		t.Fatal(err)
	}
	if err := templates.Render(testRoot); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(filepath.Join(testRoot, dest))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expectedContent {
		t.Fatal("Unexpected content: ", content)
	}
}

func TestFilterEnv(t *testing.T) {
	keep := []string{"FOO", "PATH"}
	os.Setenv("FOO", "bar")
	os.Setenv("PATH", "/bin:/usr/bin")
	os.Setenv("FILTERME", "gone")
	envs := getFilteredEnv(keep)

	if !isIn("FOO=bar", envs) || !isIn("PATH=/bin:/usr/bin", envs) || isIn("FILTERME=gone", envs) {
		t.Fatal("Unexpected env %#v", envs)
	}
}

func isIn(str string, list []string) bool {
	for _, i := range list {
		if i == str {
			return true
		}
	}
	return false
}
