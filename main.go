package main

import (
	"encoding/base64"
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
	DEF_PORT = "3000"
	DEF_MSG_404 = "Bad username/password, try again <3"
	DEF_MSG_401 = "Sorry mate, can't find this page ://"
)

var (
	PORT = DEF_PORT
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

func init() {
	log.Init(log.NewLoggerPrint())
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
		if serveFile(w, fmt.Sprint(status) + ".html") != nil {
			w.Write([]byte(def))
		}
		return
	}

	w.Write([]byte(msg))
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.CleanPath, middleware.GetHead, middleware.StripSlashes, cors.AllowAll().Handler)

	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		pathParts := prepPath(r.URL.Path)
		auth := authRoot.Resolve(pathParts)

		if auth != nil {
			h := w.Header()
			headerThing := "Basic realm=" + auth.Realm

			h.Set("WWW-Authenticate", headerThing)
			
			reqAuth := r.Header.Get("Authorization")
			authSuccess := false

			if reqAuth != "" {
				authS := strings.Split(reqAuth, " ")
				if len(authS) == 2 {
					b, err := base64.StdEncoding.DecodeString(authS[1])
					if err == nil {
						spl := strings.Split(string(b), ":")
						if len(spl) == 2 && spl[0] == auth.Name && spl[1] == auth.Password {
							authSuccess = true
						}
					}
				}
			}

			if !authSuccess {
				writeMsg(w, 401, MSG_401, DEF_MSG_401)
				return
			}
		}

		info, err := os.Stat("site" + r.URL.Path)
		if err == nil {
			if info.IsDir() {
				info, err := os.Stat("site" + r.URL.Path + "/index.html")
				if err == nil && !info.IsDir() {
					serveFile(w, "site" + r.URL.Path + "/index.html")
					return
				}
			} else {
				serveFile(w, "site" + r.URL.Path)
				return
			}
		}

		info, err = os.Stat("site" + r.URL.Path + ".html")
		if err == nil && !info.IsDir() {
			serveFile(w, "site" + r.URL.Path + ".html")
			return
		}

		writeMsg(w, 404, MSG_404, DEF_MSG_404)
	})

	log.Success("Starting server on :%s", PORT)
	http.ListenAndServe(":"+PORT, r)
}
