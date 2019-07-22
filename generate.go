package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	dataYml   = filepath.Join("assets", "data.yml")
	indexTpl  = filepath.Join("assets", "index.template")
	defsTpl   = filepath.Join("assets", "defs.template")
	indexHTML = filepath.Join("docs", "index.html")
)

// MapSlice is a wrapper of yaml.MapSlice
type MapSlice yaml.MapSlice

// HasKey checks whether the slice contains a key named `name`
func (ms MapSlice) HasKey(name string) bool {
	for _, mi := range ms {
		if key, ok := mi.Key.(string); ok && key == name {
			return true
		}
	}
	return false
}

// Find returns a value of a key `key` if exists.
// Otherwise, it throws an error.
func (ms MapSlice) Find(key string) (interface{}, error) {
	for _, mi := range ms {
		if k, ok := mi.Key.(string); ok && k == key {
			return mi.Value, nil
		}
	}
	return nil, errors.New("MapSlice.Find: cannot find a key '" + key + "'")
}

// IsSingleValue returns false if a value of a key `key` is a slice or an array.
// If it's not, then returns true.
// If the key doesn't exist, IsSingleValue throws an error.
func (ms MapSlice) IsSingleValue(key string) (bool, error) {
	val, err := ms.Find(key)
	if err != nil {
		return false, errors.Wrap(err, "MapSlice.IsSingleValue: ")
	}
	if k := reflect.ValueOf(val).Kind(); k == reflect.Slice || k == reflect.Array {
		return false, nil
	}
	return true, nil
}

// DB represents the structure of `data.yml`.
type DB map[string]MapSlice

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

	output, err := os.Create(indexHTML)
	check(err)
	defer output.Close()

	funcs := template.FuncMap{
		"noescape": func(html string) template.HTML {
			return template.HTML(html)
		},
		// The following function is originally written by tux21b on StackOverflow
		// https://stackoverflow.com/a/18276968/5989200
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("dict: number of arguments must be even")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict: keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
	tpl, err := template.New(filepath.Base(indexTpl)).Funcs(funcs).ParseFiles(indexTpl, defsTpl)
	check(err)
	err = tpl.Execute(output, db)
	check(err)
}
