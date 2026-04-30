package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	"github.com/samael0119/RoutePeek/internal/diagnostic"
	"github.com/samael0119/RoutePeek/internal/discovery"
	"github.com/samael0119/RoutePeek/internal/overview"
	"github.com/samael0119/RoutePeek/pkg/netinfo"
	"github.com/samael0119/RoutePeek/pkg/types"
)

//go:embed static/*
var staticFiles embed.FS

var (
	port               = "8080"
	openBrowser        = true
	traceRoute         = netinfo.TraceRoute
	getNetworkSnapshot = discovery.GetNetworkSnapshot
	lookupIPLocation   = fetchIPLocation
)

func main() {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/snapshot", handleSnapshot).Methods("GET")
	r.HandleFunc("/api/diagnosis", handleDiagnosis).Methods("GET")
	r.HandleFunc("/api/overview", handleOverview).Methods("GET")
	r.HandleFunc("/api/interfaces", handleInterfaces).Methods("GET")
	r.HandleFunc("/api/routes", handleRoutes).Methods("GET")
	r.HandleFunc("/api/trace", handleTrace).Methods("GET")
	r.HandleFunc("/api/ip-location", handleIPLocation).Methods("GET")
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/favicon.png")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	})

	// Serve static files from embedded FS
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))
	// Serve index.html at root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			data, err := staticFiles.ReadFile("static/dist/index.html")
			if err != nil {
				writeJSONError(w, http.StatusNotFound, "not_found", "index.html not found")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
		} else {
			http.NotFound(w, r)
		}
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("🌐 RoutePeek Web Server starting on http://localhost:%s\n", port)
	fmt.Printf("📊 Open http://localhost:%s in your browser\n", port)

	if openBrowser {
		go openBrowserFunc(fmt.Sprintf("http://localhost:%s", port))
	}

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
	}

	log.Fatal(srv.ListenAndServe())
}

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := getNetworkSnapshot()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "snapshot_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func handleDiagnosis(w http.ResponseWriter, r *http.Request) {
	snapshot, err := getNetworkSnapshot()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "diagnosis_failed", err.Error())
		return
	}

	report := diagnostic.RunDiagnosticsWithLang(snapshot, r.URL.Query().Get("lang"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func handleOverview(w http.ResponseWriter, r *http.Request) {
	snapshot, err := getNetworkSnapshot()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "overview_failed", err.Error())
		return
	}

	lang := r.URL.Query().Get("lang")
	report := diagnostic.RunDiagnosticsWithLang(snapshot, lang)
	response := overview.BuildWithLang(snapshot, report, lang)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleInterfaces(w http.ResponseWriter, r *http.Request) {
	interfaces, err := discovery.GetInterfaces()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "interfaces_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interfaces)
}

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := discovery.GetRoutes()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "routes_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func handleTrace(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		writeJSONError(w, http.StatusBadRequest, "missing_target", "target parameter is required")
		return
	}
	if !isValidTraceTarget(target) {
		writeJSONError(w, http.StatusBadRequest, "invalid_target", "target must be an IP address or domain name")
		return
	}

	hops, err := traceRoute(target)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "trace_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hops)
}

func handleIPLocation(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	if net.ParseIP(ip) == nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_ip", "ip parameter must be a valid IP address")
		return
	}

	location, err := lookupIPLocation(ip)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "ip_location_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

func fetchIPLocation(ip string) (*types.IPLocation, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,org,message", ip)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip location service returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status  string `json:"status"`
		Country string `json:"country"`
		City    string `json:"city"`
		Org     string `json:"org"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "success" {
		if payload.Message != "" {
			return nil, fmt.Errorf(payload.Message)
		}
		return nil, fmt.Errorf("ip location lookup failed")
	}
	return &types.IPLocation{
		Country: payload.Country,
		City:    payload.City,
		Org:     payload.Org,
	}, nil
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]map[string]string{
		"error": {
			"code":    code,
			"message": message,
		},
	})
}

func isValidTraceTarget(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" || len(target) > 253 {
		return false
	}
	if net.ParseIP(target) != nil {
		return true
	}

	target = strings.TrimSuffix(target, ".")
	if target == "" {
		return false
	}
	for _, label := range strings.Split(target, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		for i, r := range label {
			valid := r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-')
			if !valid {
				return false
			}
			if (i == 0 || i == len(label)-1) && r == '-' {
				return false
			}
		}
	}
	return true
}

func openBrowserFunc(url string) {
	time.Sleep(500 * time.Millisecond)

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// StartWebServer starts the web server (called from CLI)
func StartWebServer(p string, open bool) {
	port = p
	openBrowser = open

	// Change to server directory context
	os.Chdir("/app")
	main()
}
