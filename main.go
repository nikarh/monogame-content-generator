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

type ConfigFile struct {
	Templates   []string
	Content     map[string]map[string]interface{}
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


	for pattern, _ := range config.Content {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}
		config.Content[pattern]["files"] = matches

		for _, matchedFile := range matches {
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

	for _, templateFile := range config.Templates {
		templateBytes, err := ioutil.ReadFile(templateFile)
		if err != nil {
			panic(err)
		}
		parsedTemplate, err := template.New(templateFile).Parse(string(templateBytes))
		if err != nil {
			panic(err)
		}

		newFileName := strings.TrimSuffix(templateFile, filepath.Ext(templateFile))
		interpolated, err := os.OpenFile(newFileName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}

		err = parsedTemplate.Execute(interpolated, templateData)
		if err != nil {
			panic(err)
		}
	}
}
