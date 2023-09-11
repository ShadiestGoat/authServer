package main

import (
	"os"
	"path"
	"strings"

	"github.com/shadiestgoat/log"
)

type Auth struct {
	Name     string
	Password string
	Realm    string
}

type Path struct {
	A        *Auth            `json:"auth,omitempty"`
	Parent   *Path            `json:"-"`
	Children map[string]*Path `json:"children,omitempty"`
}

var authRoot = &Path{
	Children: map[string]*Path{},
}

func (p *Path) Add(part string) *Path {
	if curPath := p.Children[part]; curPath == nil {
		p.Children[part] = &Path{
			Children: map[string]*Path{},
			Parent:   p,
		}
	}

	return p.Children[part]
}

func init() {
	f, _ := os.ReadFile(".passwords")

	log.Debug("Loading config...")

	for _, l := range strings.Split(string(f), "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		segments := strings.SplitN(l, " : ", 4)
		if len(segments) < 2 || len(segments) > 3 {
			log.Fatal("Needs to have 2 or 3 segments, got '%s' (%d)", l, len(segments))
		}

		if len(segments) == 2 {
			segments = append(segments, "")
		}

		cleanPath := path.Clean(segments[0])

		if cleanPath == "." {
			cleanPath = "/"
		}

		pathSpl := strings.Split(cleanPath, "/")[1:]
		curPath := authRoot

		for _, p := range pathSpl {
			if p == "" {
				continue
			}
			if p == ".." {
				curPath = curPath.Parent
			} else if p != "." {
				curPath = curPath.Add(p)
			}
		}

		curPath.A = &Auth{
			Name:     segments[1],
			Password: segments[2],
			Realm:    cleanPath,
		}
	}

	log.Success("Loaded config!")
}
