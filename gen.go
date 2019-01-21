package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/godbus/dbus/introspect"
)

var knownTypes = make(map[string]string)

type tpMember struct {
	Type      string `xml:"type,attr"`
	Name      string `xml:"name,attr"`
	Docstring string `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 docstring"`
}

type tpStruct struct {
	Name      string     `xml:"name,attr"`
	Docstring string     `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 docstring"`
	Members   []tpMember `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 member"`
}

func (s *tpStruct) String() string {
	o := ""
	o += "// "
	o += s.Docstring
	o += "\n"
	o += "type "
	o += s.Name
	o += " struct {\n"
	for _, m := range s.Members {
		o += "\t"
		o += "// "
		o += m.Docstring
		o += "\n"
		o += "\t"
		o += m.Name
		o += " "
		o += formatType(m.Type)
		o += "\n"
	}
	return o
}

type tpMapping struct {
	Name      string     `xml:"name,attr"`
	Docstring string     `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 docstring"`
	Members   []tpMember `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 member"`
}

func (m *tpMapping) String() string {
	if len(m.Members) != 2 {
		log.Fatal("invalid number of mapping elements")
	}
	o := ""
	o += "// "
	o += m.Docstring
	o += "\ntype "
	o += m.Name
	o += " map["
	o += formatType(m.Members[0].Type)
	o += "]"
	o += formatType(m.Members[1].Type)
	return o
}

type tpSpec struct {
	XMLName  xml.Name          `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 spec"`
	Title    string            `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 title"`
	Version  string            `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 version"`
	Structs  []tpStruct        `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 struct"`
	Mappings []tpMapping       `xml:"http://telepathy.freedesktop.org/wiki/DbusSpec#extensions-v0 mapping"`
	Nodes    []introspect.Node `xml:"node"`
}

func ParseXMLSpec(data []byte) (*tpSpec, error) {
	spec := &tpSpec{}
	err := xml.Unmarshal(data, spec)
	if err != nil {
		return nil, err
	}
	for _, s := range spec.Structs {
		knownTypes[s.Name] = s.Name
	}
	for _, m := range spec.Mappings {
		knownTypes[m.Name] = m.Name
	}
	return spec, nil
}

func formatType(dbus string) string {
	if t, ok := knownTypes[dbus]; ok {
		return t
	}
	t, err := ConvertType(dbus)
	if err != nil {
		panic(err)
	}
	return t
}

func formatArgs(args ...introspect.Arg) string {
	output := ""
	for i, p := range args {
		output += p.Name
		output += " "
		output += formatType(p.Type)
		if i < len(args)-1 {
			output += ", "
		}
	}
	return output
}

func generateMethod(method introspect.Method) string {
	in := make([]introspect.Arg, 0)
	out := make([]introspect.Arg, 0)
	for _, arg := range method.Args {
		if arg.Direction == "in" {
			in = append(in, arg)
		} else {
			out = append(out, arg)
		}
	}
	output := method.Name
	output += "("
	output += formatArgs(in...)
	output += ")"
	if len(out) != 0 {
		output += " ("
		output += formatArgs(out...)
		output += ")"
	}
	return output
}

func generateInterface(iface introspect.Interface) string {
	output := ""
	output += "type "
	output += iface.Name
	output += " interface {\n"
	for _, m := range iface.Methods {
		output += "\t"
		output += generateMethod(m)
		output += "\n"
	}
	output += "}"
	return output
}

func main() {
	data, err := ioutil.ReadFile("dbus-api.xml")
	if err != nil {
		log.Fatal(err)
	}
	intro, err := ParseXMLSpec(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", intro)
	for _, s := range intro.Structs {
		fmt.Println(s.String())
	}
	for _, m := range intro.Mappings {
		fmt.Println(m.String())
	}
	for _, node := range intro.Nodes {
		for _, iface := range node.Interfaces {
			fmt.Println(generateInterface(iface))
		}
	}
}
