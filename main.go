package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	path := flag.String("path", ".", "path to serve the files from")
	basic := flag.Bool("basic", false, "enable Basic Authentication, requires you to set USERNAME and PASSWORD environment variables")
	flag.Parse()

	fs := http.FileServer(http.Dir(*path))
	if *basic {
		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")
		if username == "" || password == "" {
			log.Fatal("expected USERNAME and PASSWORD enviornment variables to be set")
		}
		http.HandleFunc("/", basicAuth(handler(fs), username, password))
	} else {
		http.HandleFunc("/", handler(fs))
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handler(fileHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS, CONNECT, PUT, DAV, dav")
			w.Header().Set("X-Api-Access-Type", "file")
			w.Header().Set("DAV", "tw5/put")

		case http.MethodGet:
			fileHandler.ServeHTTP(w, r)

		case http.MethodPost:
			if err := r.ParseMultipartForm(0); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			file, meta, err := r.FormFile("file")
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()
			f, err := os.OpenFile(meta.Filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()
			if _, err := io.Copy(f, file); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		case http.MethodPut:
			file := strings.Trim(r.URL.Path, "/")
			if file == "" {
				file = "index.html"
			}
			f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				log.Println(err)
				return
			}
			if _, err := io.Copy(f, r.Body); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
		log.Println(r.Method, r.URL)
	}
}

func basicAuth(next http.HandlerFunc, username string, password string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqUsername, reqPassword, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		usernameHash := sha256.Sum256([]byte(reqUsername))
		passwordHash := sha256.Sum256([]byte(reqPassword))
		expectedUsernameHash := sha256.Sum256([]byte(username))
		expectedPasswordHash := sha256.Sum256([]byte(password))

		// ConstantTimeCompare is use to avoid leaking information using timing attacks
		usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)
		if usernameMatch && passwordMatch {
			next.ServeHTTP(w, r)
			return
		}

	})
}
