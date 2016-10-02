package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"rsc.io/letsencrypt"
)

func loggedHandler(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s request [%s] from %s\n", prefix, r.Method, r.URL, r.RemoteAddr)
		h.ServeHTTP(w, r)
		return
	})
}

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

func startListenHTTP(conf *config) {
	if conf.HTTPLAddr != "" {
		mux := http.NewServeMux()
		for _, dir := range conf.PublicDirs {
			mux.Handle(dir.HTTPPath, http.StripPrefix(dir.HTTPPath, http.FileServer(http.Dir(dir.DirPath))))
		}
		var h http.Handler
		if conf.Logging {
			h = loggedHandler("http", mux)
		} else {
			h = mux
		}
		go func() {
			if err := http.ListenAndServe(conf.HTTPLAddr, h); err != nil {
				log.Fatalf("listening http on %s error: %s\n", conf.HTTPLAddr, err.Error())
			}
		}()
	}
}

func startListenHTTPS(conf *config) {
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

		var h http.Handler
		if conf.Logging {
			h = loggedHandler("https", mux)
		} else {
			h = mux
		}

		if conf.TLSCertPaths != nil {
			go func() {
				if err := http.ListenAndServeTLS(conf.HTTPSLAddr, conf.TLSCertPaths.TLSCertPath, conf.TLSCertPaths.TLSKeyPath, h); err != nil {
					log.Fatalf("listening https on %s error: %s\n", conf.HTTPSLAddr, err.Error())
				}
			}()
			return
		} else if conf.LetsencryptCacheFile != nil {
			var m letsencrypt.Manager
			if err := m.CacheFile(*conf.LetsencryptCacheFile); err != nil {
				log.Fatalf("setting CacheFile failed: %v\n", err)
			}
			if len(conf.Hosts) > 0 {
				m.SetHosts(conf.Hosts)
			}
			srv := &http.Server{
				Addr: conf.HTTPSLAddr,
				TLSConfig: &tls.Config{
					GetCertificate: m.GetCertificate,
				},
				Handler: h,
			}
			go func() {
				if err := srv.ListenAndServeTLS("", ""); err != nil {
					log.Fatalf("listening https on %s error: %s\n", conf.HTTPSLAddr, err.Error())
				}
			}()
			return
		} else {
			log.Printf("HTTPS is disabled because neither letsencrypt_cache_file or tls_cert_paths is set\n")
		}
	}
	log.Println("all authenticated dirs are disabled since HTTPS is disabled")
}

func serve(conf *config) {
	startListenHTTP(conf)
	startListenHTTPS(conf)
	select {}
}
