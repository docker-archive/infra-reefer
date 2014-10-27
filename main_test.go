package main

import (
	"io/ioutil"
	"os"
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
	t.Log("testRoot", testRoot)
	if err := templates.Render(testRoot); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(testRoot + dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expectedContent {
		t.Fatal("Unexpected content: ", content)
	}
}
