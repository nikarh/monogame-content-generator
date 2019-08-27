package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Template struct {
	Path string
	Body string
}

type ContentEntry struct {
	Paths     []string
	Content   []string
	FileNames []string
}

type ConfigFile struct {
	Templates   []Template
	Content     []ContentEntry
	ContentPath string `yaml:"contentPath"`
}

type ClassField struct {
	Name  string
	Value string
}

var badChars, _ = regexp.Compile(`[^\p{Ll}\p{Lu}\p{Lt}\p{Lo}\p{Nd}\p{Nl}\p{Mn}\p{Mc}\p{Cf}\p{Pc}\p{Lm}]`)

func main() {
	configFile := flag.String("config", "", "A path to a yaml config file")

	flag.Parse()

	if configFile == nil || *configFile == "" {
		println("You must pass -config argument")
		return
	}

	configBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic(err)
	}

	var config ConfigFile
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		panic(err)
	}

	var classes = make(map[string][]ClassField)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(filepath.Join(filepath.Dir(*configFile), config.ContentPath))

	for i, contentEntry := range config.Content {
		var fileNames = make([]string, 0, 0)
		for _, filePattern := range contentEntry.Paths {
			matches, err := filepath.Glob(filePattern)
			if err != nil {
				panic(err)
			}
			for _, foundFile := range matches {
				fileNames = append(fileNames, foundFile)
			}
		}

		config.Content[i].FileNames = fileNames

		for _, matchedFile := range fileNames {
			var dir = filepath.Dir(matchedFile)
			var segments = strings.Split(dir, string(os.PathSeparator))
			for i := range segments {
				segments[i] = strings.Title(segments[i])
			}
			var className = strings.Join(segments, "")
			var fieldName = strings.Title(badChars.ReplaceAllString(strings.SplitN(filepath.Base(matchedFile), ".", 2)[0], ""))

			field := ClassField{
				Name:  fieldName,
				Value: strings.TrimSuffix(matchedFile, filepath.Ext(matchedFile)),
			}

			if _, ok := classes[className]; !ok {
				classes[className] = []ClassField{field}
			} else {
				classes[className] = append(classes[className], field)
			}
		}
	}

	var templateData = make(map[string]interface{})
	templateData["groups"] = config.Content
	templateData["classes"] = classes

	_ = os.Chdir(oldWd)
	_ = os.Chdir(filepath.Dir(*configFile))

	for _, tpl := range config.Templates {
		parsedTemplate, err := template.New(tpl.Path).Parse(tpl.Body)
		if err != nil {
			panic(err)
		}

		interpolated, err := os.OpenFile(tpl.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}

		err = parsedTemplate.Execute(interpolated, templateData)
		if err != nil {
			panic(err)
		}
	}
}
