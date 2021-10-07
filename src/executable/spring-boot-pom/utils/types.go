package utils

import (
	"encoding/xml"
	"html"
	"log"
)

const (
	xmlVersion        = `<?xml version="1.0" encoding="UTF-8"?>`
	xmlns             = "http://maven.apache.org/POM/4.0.0,attr"
	xmlnsXsi          = "http://www.w3.org/2001/XMLSchema-instance"
	xmlSchemaLocation = "http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd"
)

type coordinate struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version,omitempty"`
}

type dependency struct {
	XMLName xml.Name `xml:"dependency"`
	coordinate
	Type string `xml:"pom,omitempty"`
}

type extension struct {
	XMLName xml.Name `xml:"extension"`
	coordinate
}

type plugin struct {
	XMLName xml.Name `xml:"plugin"`
}

type parent struct {
	XMLName xml.Name `xml:"parent"`
	coordinate
}

type project struct {
	XMLName xml.Name `xml:"project"`

	Xmlns             string `xml:"xmlns,attr"`
	XmlnsXsi          string `xml:"xmlns:xsi,attr"`
	XsiSchemaLocation string `xml:"xsi:schemaLocation,attr"`
	ModelVersion      string `xml:"movelVersion"`
	Properties        *properties
	Parent            *parent

	coordinate

	Dependencies []*dependency `xml:"dependencies"`
	Extensions   []*extension  `xml:"extensions"`
	Plugins      []*plugin     `xml:"plugins"`
}

func NewProject(name string) *project {
	p := &project{Xmlns: xmlns, XmlnsXsi: xmlnsXsi, XsiSchemaLocation: xmlSchemaLocation, ModelVersion: "4.0.0"}
	p.GroupId = "org.example"
	p.ArtifactId = name
	p.Version = "1.0-SNAPSHOT"

	p.Properties = NewProperties()

	return p
}

func NewDependency(groupId, artifactId, version string) *dependency {
	d := &dependency{}
	d.GroupId = groupId
	d.ArtifactId = artifactId
	d.Version = version
	return d
}

func NewExtension(groupId, artifactId, version string) *extension {
	d := &extension{}
	d.GroupId = groupId
	d.ArtifactId = artifactId
	d.Version = version
	return d
}

func (p *project) AddDependency(d *dependency) {
	p.Dependencies = append(p.Dependencies, d)
}

func (p *project) AddDependencyByID(groupId string, artifactId string, version string) {
	p.AddDependency(NewDependency(groupId, artifactId, version))
}

func (p *project) AddDependencies(ds ...*dependency) {
	for _, d := range ds {
		p.AddDependencies(d)
	}
}

func (p *project) AddParent(groupId GroupIDType) {
	p.Parent = &parent{}
	p.Parent.GroupId = groupIDMap[groupId]
	p.Parent.ArtifactId = GetParentArtifactId(groupId)
	p.Parent.Version = GetParentVersion(groupId)
}

func (p *project) ToXML() string {
	b, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	return html.UnescapeString(xmlVersion + "\n" + string(b))
}
