package main

import (
	"fmt"
	"log"
	"net/http"
)

func authedHandler(realm string, users map[string]string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqUser, reqPass, ok := r.BasicAuth(); ok {
			if pass, ok := users[reqUser]; ok && pass == reqPass {
				h.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%s", realm))
		w.WriteHeader(http.StatusUnauthorized)
		return
	})
}

func allowedUsers(allowedUsernames []string, allUsers map[string]string) (users map[string]string) {
	users = make(map[string]string)
	for _, username := range allowedUsernames {
		if password, ok := allUsers[username]; ok {
			users[username] = password
		} else {
			log.Fatalf("username %s is not found.", username)
		}
	}
	return
}

func serve(conf *config) {
	var err error

	if conf.HTTPLAddr != "" {
		mux := http.NewServeMux()
		for _, dir := range conf.PublicDirs {
			mux.Handle(dir.HTTPPath, http.StripPrefix(dir.HTTPPath, http.FileServer(http.Dir(dir.DirPath))))
		}
		go func() {
			if err = http.ListenAndServe(conf.HTTPLAddr, mux); err != nil {
				log.Fatalf("listening http on %s error: %s\n", conf.HTTPLAddr, err.Error())
			}
		}()
	}

	if conf.HTTPSLAddr != "" {
		mux := http.NewServeMux()
		for _, dir := range conf.PublicDirs {
			mux.Handle(dir.HTTPPath, http.StripPrefix(dir.HTTPPath, http.FileServer(http.Dir(dir.DirPath))))
		}
		for _, dir := range conf.AuthenticatedDirs {
			mux.Handle(dir.HTTPPath, authedHandler(
				dir.HTTPPath, allowedUsers(dir.Usernames, conf.Users),
				http.StripPrefix(dir.HTTPPath, http.FileServer(http.Dir(dir.DirPath)))),
			)
		}
		go func() {
			if err = http.ListenAndServeTLS(conf.HTTPSLAddr, conf.TLSCertPath, conf.TLSKeyPath, mux); err != nil {
				log.Fatalf("listening https on %s error: %s\n", conf.HTTPSLAddr, err.Error())
			}
		}()
	} else {
		log.Println("since HTTPS is not configured, all authenticated dirs are disabled.")
	}

	select {}
}
