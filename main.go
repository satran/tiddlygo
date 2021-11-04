package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	http.HandleFunc("/", basicAuth(handler(fs), username, password))
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
