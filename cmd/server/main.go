package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ultima-pista/ultima-pista/internal/api"
	"github.com/ultima-pista/ultima-pista/internal/store"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "endereço HTTP")
	dbPath := flag.String("db", envOr("DB_PATH", "data/ultima-pista.db"), "caminho do SQLite")
	static := flag.String("static", envOr("STATIC_DIR", "web/dist"), "frontend compilado")
	secure := flag.Bool("secure", envOr("SECURE_COOKIES", "") == "1", "cookies Secure")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		log.Fatal(err)
	}
	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	staticDir := *static
	if _, err := os.Stat(staticDir); err != nil {
		log.Printf("aviso: static dir %s indisponível (%v); API-only", staticDir, err)
		staticDir = ""
	}

	srv := api.New(st, *secure, staticDir)
	server := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("Última Pista ouvindo em %s (db=%s)", *addr, *dbPath)
	log.Fatal(server.ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
