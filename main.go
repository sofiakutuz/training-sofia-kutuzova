// Služba objednávky: přijímá objednávky po HTTP a ukládá je do souboru.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	listen := flag.String("listen", ":8080", "adresa, na které služba poslouchá")
	data := flag.String("data", "/data", "adresář, kam se ukládají objednávky")
	flag.Parse()
	soubor := filepath.Join(*data, "objednavky.txt")
	if os.Getenv("DB_PASSWORD") == "" {
		log.Print("varování: DB_PASSWORD není nastavené, objednávky se nespárují s databází")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/objednavky", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			f, err := os.OpenFile(soubor, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				log.Printf("objednávku nejde uložit: %v", err)
				http.Error(w, "objednávku nejde uložit", http.StatusInternalServerError)
				return
			}
			defer f.Close()
			fmt.Fprintf(f, "%s objednávka\n", time.Now().Format(time.RFC3339))
			fmt.Fprintln(w, "objednávka uložena")
			return
		}
		obsah, _ := os.ReadFile(soubor)
		w.Write(obsah)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "objednavky běží na %s\n", hostname())
		if sklad := os.Getenv("SKLAD_URL"); sklad != "" {
			fmt.Fprintf(w, "sklad %s: %s\n", sklad, stavSkladu(sklad))
		}
	})

	srv := &http.Server{Addr: *listen, Handler: mux}
	go func() {
		log.Printf("poslouchám na %s", *listen)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Na SIGTERM (docker stop, Kubernetes) dokončí rozpracované requesty a skončí.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	s := <-stop
	log.Printf("dostal jsem %v, dokončuji rozpracované požadavky", s)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Print("ukončeno")
}

func hostname() string {
	h, _ := os.Hostname()
	return h
}

func stavSkladu(url string) string {
	c := http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get(url + "/healthz")
	if err != nil {
		return "nedostupný (" + err.Error() + ")"
	}
	resp.Body.Close()
	return "dostupný"
}
