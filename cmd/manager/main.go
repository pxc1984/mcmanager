package main

import (
	"log"
	"net/http"

	"mcmanager/internal/config"
	"mcmanager/internal/manager"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	mgr := manager.New(cfg)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/update", mgr.UpdateHandler)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, nil); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
