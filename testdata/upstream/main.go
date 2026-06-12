// go run ./testdata/upstream
// # listens on :3000
// go run ./testdata/upstream -addr :3001
// seulement pour demarrer les test en go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":3000", "listen address")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from upstream %s — %s %s\n", *addr, r.Method, r.URL.Path)
	})

	// Useful later: test timeouts and retries.
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "slow response")
	})

	// Useful later: test the circuit breaker.
	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	log.Printf("dummy upstream listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
