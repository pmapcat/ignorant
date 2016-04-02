package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/hoisie/mustache"
	"github.com/kardianos/osext"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var TEMPLATE = `
##############################################
This is a composed .gitignore file for
{{#Items}}{{LangName}} {{/Items}}
This file is MIT licensed.
It uses https://github.com/github/gitignore repo 
for composable components.

##############################################
{{#Items}}
###### Gitignore file for {{LangName}} ######
###### Source: {{Url}} ######
{{{Src}}} 
{{/Items}}
######################################
`

const (
	FETCH_URL = "https://github.com/github/gitignore"
)

var EXEC_DIR, _ = osext.ExecutableFolder()
var GITIGNORE_DIR = filepath.Join(EXEC_DIR, "gitignore")

type ResultDatum struct {
	Path string
	Src  string
}

func (r ResultDatum) LangName() string {
	return strings.ToLower(strings.Replace(filepath.Base(r.Path), filepath.Ext(r.Path), "", -1))
}

func (r ResultDatum) Url() string {
	return fmt.Sprintf("%s/blob/master/%s", FETCH_URL, filepath.Base(r.Path))
}

func FetchRepo() {
	// if does not exist, clone
	if _, err := os.Stat(GITIGNORE_DIR); os.IsNotExist(err) {
		fmt.Sprintf("Fetching gitignore repo from: %s", FETCH_URL)
		data, err := exec.Command("git", "-C", GITIGNORE_DIR, "clone", FETCH_URL).CombinedOutput()
		if err != nil {
			fmt.Println(string(data), err)
		} else {
			fmt.Println(string(data))
		}
	}
	// if exists,pull
	if _, err := os.Stat(GITIGNORE_DIR); err == nil {
		fmt.Printf("Updating gitnigore repo from: %s", FETCH_URL)
		data, err := exec.Command("git", "-C", GITIGNORE_DIR, "pull", FETCH_URL).CombinedOutput()
		if err != nil {
			fmt.Println(string(data), err)
		} else {
			fmt.Println(string(data))
		}
	}
}

func ShowPossibleIgnores(data map[string]ResultDatum) {
	for _, v := range data {
		fmt.Printf(v.LangName() + "\n")
	}
}

func GetPossibleIgnores() map[string]ResultDatum {
	result := map[string]ResultDatum{}
	filepath.Walk(GITIGNORE_DIR, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(info.Name()) == ".gitignore" {
			item := ResultDatum{}
			item.Path = curPath
			data, err := ioutil.ReadFile(curPath)
			if err == nil {
				item.Src = string(data)
			} else {
				return err
			}
			result[item.LangName()] = item
		}
		return nil
	})
	return result
}

func Compose(isStdout bool, data map[string]ResultDatum) string {
	// make it composable
	composable := []ResultDatum{}
	for _, v := range data {
		composable = append(composable, v)
	}
	result := mustache.Render(TEMPLATE, map[string][]ResultDatum{"Items": composable})
	if isStdout {
		return result
	} else {
		fmt.Println(ioutil.WriteFile(".gitignore", []byte(result), 0777))
		return "Written to .gitignore file. Check it out"
	}
}

func RunAll(isStdout bool, args []string) string {
	ignores := GetPossibleIgnores()
	langs := map[string]ResultDatum{}
	for _, v := range args {
		curName := strings.ToLower(v)
		if val, ok := ignores[curName]; ok {
			langs[curName] = val
		}
	}
	return Compose(isStdout, langs)
}

func main() {
	app := cli.NewApp()
	isBool := false
	app.Name = "Ignorant"
	app.Usage = "Easily compose different .gitignore files"
	app.Action = func(c *cli.Context) {
		println("Please, specify gitignore files you want to include")
	}
	app.Version = "0.1.13"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "stdout",
			Destination: &isBool,
			Usage:       "should I output .gitignore to stdout",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "use",
			Aliases: []string{"u"},
			Usage:   "select gitignore names to use, for example: Leiningen Clojure",
			Action: func(c *cli.Context) {
				FetchRepo()
				fmt.Print(RunAll(isBool, c.Args()))
			},
		},
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "show possible gitignores",
			Action: func(c *cli.Context) {
				FetchRepo()
				ShowPossibleIgnores(GetPossibleIgnores())
			},
		},
	}
	app.Run(os.Args)
}
