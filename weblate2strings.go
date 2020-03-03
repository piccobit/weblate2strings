package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/alecthomas/kong"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type Context struct {
	Verbose bool
}

type VersionCmd struct {}

type YamlCmd struct {
	InputPattern string `arg name:"inputPattern" help:"Path pattern to the weblate input files." type:"string"`
	OutputDir string `arg name:"output" help:"Path to the directory where the Monkey C output resource files are stored." type:"path"`
	WeblateContext string `arg optional name:"context" help:"Weblate context." type:"string" default:"weblate"`
}

type WeblateStrings map[string]string

type WeblateYaml map[string]WeblateStrings

var Version = "development"

var cli struct {
	Verbose bool `help:"Verbose mode."`

	Yaml YamlCmd `cmd help:"Convert from Yaml."`
	Version VersionCmd `cmd help:"Display version."`
}

func (v *VersionCmd) Run(ctx *Context) error {
	fmt.Printf("Version: %s", Version)

	return nil
}

func (y *YamlCmd) Run(ctx *Context) error {
	 weblateYaml := WeblateYaml{}

	 inputFiles, err := filepath.Glob(y.InputPattern)
	 check(err)

	 for _, inputFileName := range inputFiles {
		 inputFile, err := ioutil.ReadFile(inputFileName)
		 check(err)

		 err = yaml.Unmarshal(inputFile, &weblateYaml)
		 check(err)

		 if ctx.Verbose {
			 for k, v := range weblateYaml[y.WeblateContext] {
				 fmt.Printf("%s - %s\n", k, v)
			 }
		 }

		 re := regexp.MustCompile(`messages\.(.+)\.(yml|yaml)`)
		 languages := re.FindStringSubmatch(inputFileName)

		 if languages == nil {
		 	fmt.Printf("Did not find language code in file names.\n")
		 	os.Exit(1)
		 }

		 languageBase, _ := language.Make(languages[1]).Base()

		 languageCode := languageBase.ISO3()

		 var outputFileName string

		 if languageCode != "eng" {
			 outputFileName = filepath.Join(y.OutputDir, "resources-" + languageCode, "strings.xml")
		 } else {
			 outputFileName = filepath.Join(y.OutputDir, "resources", "strings.xml")
		 }

		 outputFile, err := os.Create(outputFileName)
		 check(err)

		 tmpl := template.Must(template.ParseFiles("strings.tmpl"))

		 err = tmpl.Execute(outputFile, weblateYaml[y.WeblateContext])
		 check(err)

		 err = outputFile.Close()
		 check(err)
	 }

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run(&Context{Verbose: cli.Verbose})
	ctx.FatalIfErrorf(err)
}
