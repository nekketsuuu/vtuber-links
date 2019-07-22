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

// dict converts (key1, value1, key2, value2, ...) into a dictionary {key1: value1, key2: value2, ...}.
func dict(values ...interface{}) (map[string]interface{}, error) {
	// The following implementation is originally written by tux21b on StackOverflow
	// https://stackoverflow.com/a/18276968/5989200
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
}

// hasKey checks whether the slice contains a key named `name`.
func hasKey(msi interface{}, name string) (bool, error) {
	ms, ok := msi.(yaml.MapSlice)
	if !ok {
		return false, errors.New("hasKey: cannot cast to yaml.MapSlice")
	}
	for _, mi := range ms {
		if key, ok := mi.Key.(string); ok && key == name {
			return true, nil
		}
	}
	return false, nil
}

// find returns a value of a key `key` if exists.
// Otherwise, it throws an error.
func find(msi interface{}, key string) (interface{}, error) {
	ms, ok := msi.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("find: cannot cast to yaml.MapSlice")
	}
	for _, mi := range ms {
		if k, ok := mi.Key.(string); ok && k == key {
			return mi.Value, nil
		}
	}
	return nil, errors.New("find: cannot find a key '" + key + "'")
}

// isSingleValue returns false if a value of a key `key` is a slice or an array.
// If it's not, then returns true.
// If the key doesn't exist, IsSingleValue throws an error.
func isSingleValue(msi interface{}, key string) (bool, error) {
	ms, ok := msi.(yaml.MapSlice)
	if !ok {
		return false, errors.New("isSingleValue: cannot cast to yaml.MapSlice")
	}
	val, err := find(ms, key)
	if err != nil {
		return false, errors.Wrap(err, "isSingleValue: cannot find a key")
	}
	if k := reflect.ValueOf(val).Kind(); k == reflect.Slice || k == reflect.Array {
		return false, nil
	}
	return true, nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	data, err := ioutil.ReadFile(dataYml)
	check(err)
	db := yaml.MapSlice{}
	err = yaml.Unmarshal(data, &db)
	check(err)

	output, err := os.Create(indexHTML)
	check(err)
	defer output.Close()

	funcs := template.FuncMap{
		"noescape": func(html string) template.HTML {
			return template.HTML(html)
		},
		"dict":          dict,
		"hasKey":        hasKey,
		"find":          find,
		"isSingleValue": isSingleValue,
	}
	tpl, err := template.New(filepath.Base(indexTpl)).Funcs(funcs).ParseFiles(indexTpl, defsTpl)
	check(err)
	err = tpl.Execute(output, db)
	check(err)
}
