package utils

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
)

type GroupIDType int
type ArtifactIDType int
type DefaultVersionType int

func (g GroupIDType) String() string {
	fmt.Println("here", groupIDMap[g], int(g))
	return groupIDMap[g]
}

func (a ArtifactIDType) String() string {
	return artifactIDMap[a]
}

const (
	Springboot = iota + 1
	SpringbootStarterWeb
)

var (
	hasParent containerW.Set
)

// data should be added here
var (
	groupIDMap = map[GroupIDType]string{
		Springboot: "org.springframework.boot",
	}

	artifactIDMap = map[ArtifactIDType]string{
		SpringbootStarterWeb: "spring-boot-starter-web",
	}

	// parent 标签 根据groupID拿到artifactID
	groupID2Artifact = map[GroupIDType]string{
		Springboot: "spring-boot-starter-parent",
	}

	// parent 标签 根据groupID拿到version
	groupID2Version = map[GroupIDType]string{
		Springboot: "2.4.0",
	}
	// reverse maps
	groupIDMapReversed    map[string]GroupIDType
	artifactIDMapReversed map[string]ArtifactIDType
)

// reverse string
func init() {
	hasParent = *containerW.NewSet()
	hasParent.Add("")
	groupIDMapReversed = make(map[string]GroupIDType)
	artifactIDMapReversed = make(map[string]ArtifactIDType)
	for k, v := range groupIDMap {
		groupIDMapReversed[v] = k
	}

	for k, v := range artifactIDMap {
		artifactIDMapReversed[v] = k
	}
}

// ---

func GetGroupId(name GroupIDType) string {
	res, ok := groupIDMap[name]
	if !ok {
		fmt.Println(color.RedString("GroupID %q nout found\n", name))
	}
	return res
}

func GetArtifactId(name ArtifactIDType) string {
	res, ok := artifactIDMap[name]
	if !ok {
		fmt.Println(color.RedString("ArtifactID %q nout found\n", name))
	}
	return res
}
