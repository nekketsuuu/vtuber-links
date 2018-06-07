package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	dataYml   = filepath.Join("assets", "data.yml")
	indexTpl  = filepath.Join("assets", "index.template")
	indexHtml = filepath.Join("docs", "index.html")
)

type (
	Site   map[string]string
	Person map[string]Site
	DB     map[string]Person
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	data, err := ioutil.ReadFile(dataYml)
	check(err)
	db := make(DB)
	err = yaml.Unmarshal(data, &db)
	check(err)

	output, err := os.Create(indexHtml)
	check(err)
	defer output.Close()

	tpl, err := template.ParseFiles(indexTpl)
	check(err)
	err = tpl.Execute(output, db)
	check(err)
}
