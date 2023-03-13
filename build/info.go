package build

import (
	"path"
	"strings"
)

//Info represent plugin info
type Info struct {
	Name    string
	Scn     int //sequence change number, time based format YYYYMMDDHHMMSS at utc
	Runtime Runtime
}

//DecodeURL decode info from URL
func (p *Info) DecodeURL(URL string) {
	name := path.Base(URL)
	name = name[:len(name)-3]
	if index := strings.LastIndex(name, "_"); index != -1 {
		p.Runtime.Arch = name[index+1:]
		name = name[:index]
	}
	if index := strings.LastIndex(name, "_"); index != -1 {
		p.Runtime.Os = name[index+1:]
		name = name[:index]
	}
	if index := strings.LastIndex(name, "_"); index != -1 {
		p.Runtime.Version = name[index+1:]
		name = name[:index]
	}
	if index := strings.LastIndex(name, "_"); index != -1 {
		p.Runtime.Version = name[index+1:] + "." + p.Runtime.Version
		name = name[:index]
	}
	if index := strings.LastIndex(name, "_"); index != -1 {
		p.Runtime.Version = name[index+1:] + "." + p.Runtime.Version
		name = name[:index]
	}
}
