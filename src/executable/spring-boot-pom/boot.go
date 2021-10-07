package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fatih/color"
	utils "github.com/grewwc/go_tools/src/executable/spring-boot-pom/utils"
	"github.com/grewwc/go_tools/src/terminalW"
)

var choices map[utils.GroupIDType][]utils.ArtifactIDType

// set up all choices
func init() {
	choices = make(map[utils.GroupIDType][]utils.ArtifactIDType)
	choices[utils.Springboot] = []utils.ArtifactIDType{
		utils.SpringbootStarterWeb, // spring-boot-starter-web
	}
}

func test() {
	p := utils.NewProject("first-demo")
	p.AddDependencyByID(utils.GetGroupId(utils.Springboot),
		utils.GetArtifactId(utils.SpringbootStarterWeb), "2.4.0")
	fmt.Println(p.ToXML())
	m := make(map[string]string)
	fmt.Println(m["gs"])
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("ls", false, "list all possible choices")
	fs.String("t", "", "the type of pom.xml")

	res := terminalW.ParseArgsCmd("ls")
	if res == nil {
		fs.PrintDefaults()
		return
	}

	if res.ContainsFlagStrict("ls") { // list all choices
		cnt := 1
		for g, as := range choices {
			for _, a := range as {
				fmt.Println(color.YellowString("%d >> \t%q\n", cnt, utils.NormGroupAndArtifactID(g, a)))
				cnt++
			}
		}
		return
	}

	t, err := res.GetFlagVal("t")
	if err != nil {
		log.Fatalln(err)
	}

	groupId, artifactId := utils.GetCoordByNormString(t)

	p := utils.NewProject("example")
	p.AddParent(groupId)
	p.AddDependencyByID(groupId.String(), artifactId.String(), "")
	fmt.Println(p.ToXML())

}
