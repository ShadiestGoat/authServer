package main

import (
	"path"
	"strings"
)

func prepPath(pathInput string) []string {
	clean := path.Clean(pathInput)
	if path.Ext(pathInput) == ".html" {
		clean = clean[:len(clean)-5]
		pathInput = pathInput[:len(pathInput)-5]
	}
	if clean == "." || clean == "/" {
		return []string{}
	}

	return strings.Split(pathInput, "/")[1:]
}

// Assumes that the input here is clean
func (p Path) Resolve(path []string) *Conf {
	if len(path) == 0 {
		return p.Conf
	}
	curPath := path[0]

	basic := p.Children[curPath]
	if basic != nil {
		if a := basic.Resolve(path[1:]); a != nil {
			return a
		}
	}

	wild1 := p.Children["*"]
	if wild1 != nil {
		if a := wild1.Resolve(path[1:]); a != nil {
			return a
		}
	}

	wild2 := p.Children["**"]
	if wild2 != nil {
		if len(path) == 1 {
			return wild2.Conf
		}

		for i, part := range path {
			r := wild2.Children[part]
			if r == nil {
				continue
			}
			a := r.Resolve(path[i+1:])
			if a != nil {
				return a
			}
		}
	}

	return nil
}
