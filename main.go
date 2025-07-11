package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var (
	port     = flag.String("l", "8087", "port to listen on")
	modelCfg = flag.String("m", "", "name_port,name_port")
)

func parseModelConfig(modelCfg string) (map[string]*httputil.ReverseProxy, error) {
	models := make(map[string]*httputil.ReverseProxy)
	for _, model := range strings.Split(modelCfg, ",") {
		parts := strings.Split(model, "_")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid model configuration: %s", model)
		}

		modelName := parts[0]
		port := parts[1]

		log.Printf("Routing %s to %s\n", modelName, port)
		models[modelName] = httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%s", port),
		})
	}
	return models, nil
}

func jsonError(w http.ResponseWriter, message string, code int) {
	log.Println(message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func main() {
	flag.Parse()

	if *modelCfg == "" {
		log.Fatal("model configuration is required")
	}
	models, err := parseModelConfig(*modelCfg)
	if err != nil {
		log.Fatal(err)
	}

	serve := func(modelName string, w http.ResponseWriter, r *http.Request) {
		proxy, found := models[modelName]
		if !found {
			jsonError(w, "model not found", http.StatusNotFound)
			return
		}

		proxy.ServeHTTP(w, r)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			jsonError(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			jsonError(w, fmt.Sprintf("failed to find model parameter in request body: %v", err), http.StatusBadRequest)
			return
		}
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		serve(body.Model, w, r)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		modelName := r.URL.Query().Get("model")
		serve(modelName, w, r)
	})

	log.Printf("Starting model router on port %s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}
