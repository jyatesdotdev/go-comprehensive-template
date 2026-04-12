// Example api demonstrates a RESTful server with CRUD endpoints and an HTTP client.
//
// Run: go run ./examples/api
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/example/go-template/internal/api"
)

func main() {
	store := api.NewStore()
	handler := api.Chain(api.ItemHandler(store), api.Recovery, api.Logging, api.CORS)

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		api.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	srv := &http.Server{Addr: ":9090", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Println("server listening on :9090")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // wait for server

	// --- Client demo ---
	base := "http://localhost:9090"
	fmt.Println("=== RESTful API Client Demo ===")

	// Create items
	for _, it := range []api.Item{{ID: "1", Name: "Alpha"}, {ID: "2", Name: "Beta"}} {
		body, _ := json.Marshal(it)
		resp, err := http.Post(base+"/items", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Fatal(err)
		}
		printResp("POST /items", resp)
	}

	// List
	resp, _ := http.Get(base + "/items")
	printResp("GET  /items", resp)

	// Get one
	resp, _ = http.Get(base + "/items/1")
	printResp("GET  /items/1", resp)

	// Delete
	req, _ := http.NewRequest(http.MethodDelete, base+"/items/1", http.NoBody)
	resp, _ = http.DefaultClient.Do(req)
	printResp("DEL  /items/1", resp)

	// 404
	resp, _ = http.Get(base + "/items/1")
	printResp("GET  /items/1", resp)

	_ = srv.Close() // #nosec G104 -- demo shutdown
	fmt.Println("\ndone")
}

func printResp(label string, resp *http.Response) {
	defer resp.Body.Close() //nolint:errcheck // best-effort close
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("%-18s %d %s", label, resp.StatusCode, b)
}
