package main

import (
	"os"
	"path"
	"strings"

	"github.com/shadiestgoat/log"
	"gopkg.in/yaml.v3"
)

type Conf struct {
	Auth       *ConfAuth         `yaml:"auth,omitempty"`
	Headers    map[string]string `yaml:"headers"`
	Redirect   string            `yaml:"redirect"`
	FakeRender string            `yaml:"fakeRender"`
}

type ConfAuth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Realm    string `yaml:"realm"`
}

type Path struct {
	Conf     *Conf            `json:"auth,omitempty"`
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

func cleanPath(p string) string {
	if p == "" {
		return p
	}
	p = path.Clean(p)

	if p == "." {
		return "/"
	}

	return p
}

func init() {
	log.Init(log.NewLoggerPrint(), log.NewLoggerFileComplex("serverLogs/log", log.FILE_DESCENDING, 5))

	f, err := os.OpenFile("config.yaml", os.O_RDONLY, 0755)
	log.FatalIfErr(err, "loading config.yaml")

	var allConfig = map[string]*Conf{}

	log.FatalIfErr(yaml.NewDecoder(f).Decode(allConfig), "parsing config.yaml")

	for fullP, c := range allConfig {
		rPath := cleanPath(fullP)

		if c.Headers == nil {
			c.Headers = map[string]string{}
		}

		if c.Auth == nil {
			continue
		}

		c.Auth.Realm = strings.TrimSpace(c.Auth.Realm)

		if c.Auth.Realm == "" {
			c.Auth.Realm = rPath
		}

		if c.FakeRender != "" {
			c.FakeRender = cleanPath(c.FakeRender)
		}

		if c.Redirect != "" {
			c.Redirect = cleanPath(c.Redirect)
		}

		pathSpl := strings.Split(rPath, "/")[1:]
		curPath := authRoot

		for _, p := range pathSpl {
			if p == "" {
				continue
			}

			curPath = curPath.Add(p)
		}

		curPath.Conf = c
	}

	log.Success("Loaded config!")
}
