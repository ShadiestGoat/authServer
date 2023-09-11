package main

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/shadiestgoat/log"
)

var PORT = "3000"

func init() {
	godotenv.Load(".env")

	if p := os.Getenv("PORT"); p != "" {
		PORT = p
	}
}

func init() {
	log.Init(log.NewLoggerPrint())
}

func serveFile(w http.ResponseWriter, path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func main() {
	r := chi.NewRouter()

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
				w.WriteHeader(401)
				w.Write([]byte("Auth failed <3"))
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
			serveFile(w, "site" + r.URL.Path + "/index.html")
			return
		}

		w.WriteHeader(404)
		w.Write([]byte("Sorry mate, can't find it ://"))
	})

	log.Success("Starting server on :%s", PORT)
	http.ListenAndServe(":"+PORT, r)
}
