package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/shadiestgoat/log"
)

const (
	DEF_PORT    = "3000"
	DEF_MSG_401 = "Bad username/password, try again <3"
	DEF_MSG_404 = "Sorry mate, can't find this page ://"
)

var (
	PORT    = DEF_PORT
	MSG_404 = DEF_MSG_404
	MSG_401 = DEF_MSG_401
)

type opt struct {
	Env string
	Opt *string
}

func init() {
	godotenv.Load(".env")

	opts := []*opt{
		{"PORT", &PORT},
		{"MSG_404", &MSG_404},
		{"MSG_401", &MSG_401},
	}

	for _, o := range opts {
		if v, ok := os.LookupEnv(o.Env); ok {
			*o.Opt = v
		}
	}
}

func serveFile(w http.ResponseWriter, filePath string) error {
	mt := mime.TypeByExtension(path.Ext(filePath))

	if mt != "" {
		w.Header().Set("Content-Type", mt)
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func writeMsg(w http.ResponseWriter, status int, msg string, def string) {
	w.WriteHeader(status)

	if msg == "<file>" {
		if serveFile(w, fmt.Sprint(status)+".html") != nil {
			w.Write([]byte(def))
		}
		return
	}

	w.Write([]byte(msg))
}

func authIsGood(h string, a *ConfAuth) bool {
	if h == "" {
		return false
	}

	authS := strings.SplitN(h, " ", 2)
	if len(authS) != 2 || authS[0] != "Basic" {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(authS[1])
	if err != nil {
		return false
	}

	spl := strings.Split(string(b), ":")
	return len(spl) == 2 && spl[0] == a.Username && spl[1] == a.Password
}

func fileIsServable(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func getPageFile(p string) string {
	p = path.Join("site", p)

	info, err := os.Stat(p)
	if err == nil {
		if info.IsDir() {
			p = path.Join(p, "index.html")
			if fileIsServable(p) {
				return p
			}
		} else {
			return p
		}
	}

	if !errors.Is(err, os.ErrNotExist) {
		log.Warn("Err while stat for '%v': %v", p, err)
	}

	p += ".html"

	if fileIsServable(p) {
		return p
	}

	return ""
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.CleanPath, middleware.GetHead, middleware.StripSlashes, cors.AllowAll().Handler)

	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		pathParts := prepPath(r.URL.Path)
		conf := authRoot.Resolve(pathParts)

		var pageToRender = r.URL.Path

		if conf != nil {
			respH := w.Header()

			for h, v := range conf.Headers {
				respH.Add(h, v)
			}

			if conf.Auth != nil {
				respH.Set("WWW-Authenticate", "Basic realm="+conf.Auth.Realm)

				if !authIsGood(r.Header.Get("Authorization"), conf.Auth) {
					writeMsg(w, 401, MSG_401, DEF_MSG_401)
					return
				}
			}

			if conf.Redirect != "" {
				http.Redirect(w, r, conf.Redirect, 307)
				return
			}

			if conf.FakeRender != "" {
				pageToRender = conf.FakeRender
			}
		}

		file := getPageFile(pageToRender)

		if file == "" {
			writeMsg(w, 404, MSG_404, DEF_MSG_404)
		}

		serveFile(w, file)
	})

	log.Success("Starting server on :%s", PORT)
	http.ListenAndServe(":"+PORT, r)
}
