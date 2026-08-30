package discover

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/clemenzi/macaron/internal/macaron/service"
)

// Port is the TCP port used by the discovery page.
const Port = 49000

//go:embed templates/*.html static/*.css
var content embed.FS

var pageTemplate = template.Must(template.ParseFS(content, "templates/*.html"))

type pageData struct {
	Services []viewEndpoint
}

type viewEndpoint struct {
	Name string
	URL  string
}

// Serve publishes running services until the context is cancelled.
func Serve(ctx context.Context, services []service.Endpoint) error {
	server := &http.Server{
		Addr:              ":" + strconv.Itoa(Port),
		Handler:           handler(services),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func handler(services []service.Endpoint) http.Handler {
	services = append([]service.Endpoint{}, services...)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		data := pageData{Services: make([]viewEndpoint, 0, len(services))}
		for _, endpoint := range services {
			data.Services = append(data.Services, viewEndpoint{
				Name: endpoint.Name,
				URL:  "http://" + net.JoinHostPort(requestHostname(r), strconv.Itoa(endpoint.Port)),
			})
		}

		var body bytes.Buffer
		if err := pageTemplate.ExecuteTemplate(&body, "index.html", data); err != nil {
			http.Error(w, "Unable to render discovery page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = body.WriteTo(w)
	})

	mux.HandleFunc("GET /api/services", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_ = json.NewEncoder(w).Encode(services)
	})

	mux.HandleFunc("GET /assets/style.css", func(w http.ResponseWriter, _ *http.Request) {
		stylesheet, err := content.ReadFile("static/style.css")
		if err != nil {
			http.Error(w, "Stylesheet unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(stylesheet)
	})

	return mux
}

func requestHostname(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		return host
	}
	return r.Host
}
