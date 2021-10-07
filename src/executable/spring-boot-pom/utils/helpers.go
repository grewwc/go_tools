package utils

import (
	"fmt"
	"strings"
)

func NormGroupAndArtifactID(g GroupIDType, a ArtifactIDType) string {
	return fmt.Sprintf("%s:%s", GetGroupId(g), GetArtifactId(a))
}

func GetCoordByNormString(norm string) (GroupIDType, ArtifactIDType) {
	data := strings.Split(norm, ":")
	groupId, artifactId := data[0], data[1]
	return groupIDMapReversed[groupId], artifactIDMapReversed[artifactId]
}

func GetParentArtifactId(groupId GroupIDType) string {
	return groupID2Artifact[groupId]
}

func GetParentVersion(groupId GroupIDType) string {
	return groupID2Version[groupId]
}
