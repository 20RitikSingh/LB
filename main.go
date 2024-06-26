package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	webui "github.com/20ritiksingh/LoadBalancer/WebUI"
	"github.com/20ritiksingh/LoadBalancer/servers"
	"github.com/20ritiksingh/LoadBalancer/test"
	"github.com/fsnotify/fsnotify"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	rr          int
)

func run() {
	go test.Test()
	go webui.Webui()
	test.StartAutoScaling(2)

	// Initialize Redis client
	redisClient = servers.Init()

	// Load server information from the YAML file
	serverListRaw, err := redisClient.LRange(context.Background(), "servers", 0, -1).Result()
	if err != nil {
		log.Fatal("Error getting server list:", err)
	}

	serverList := make([]*url.URL, len(serverListRaw))
	for i, rawURL := range serverListRaw {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			log.Fatal("Error parsing server URL:", err)
		}
		serverList[i] = parsedURL
	}
	if err != nil {
		log.Fatal("Error getting list elements:", err)
	}

	// Print server information
	fmt.Println("List elements:")
	for _, val := range serverList {
		fmt.Println("-", val)
	}

	proxy := NewStickyReverseProxy(serverList)

	server := http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}
	// server.Close()

	fmt.Println("Reverse proxy server listening on port 8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}
func main() {
	// Start watching the YAML file
	filename := "./servers/servers.yaml"
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal("Error adding file to watcher:", err)
	}

	// Initial run of the main function
	run()

	// Watch for changes in the YAML file
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("File modified:", event.Name)
					run()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Wait for signals to exit
	select {}
}

func parseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

// SessionManager manages session persistence
type SessionManager struct {
	sync.Mutex
}

// NewStickyReverseProxy creates a new reverse proxy with sticky sessions
func NewStickyReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	sm := &SessionManager{}
	var target *url.URL
	director := func(req *http.Request) {
		ip := req.RemoteAddr
		forwardedFor := req.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			// The X-Forwarded-For header may contain a comma-separated list
			// of IP addresses, with the client's IP address being the first one
			ips := strings.Split(forwardedFor, ",")
			ip = strings.TrimSpace(ips[0])
		}

		if ip != "" {
			targetURL, err := redisClient.HGet(req.Context(), "sessions", ip).Result()
			if err == nil && targetURL != "" {
				parsedURL, _ := url.Parse(targetURL)
				req.URL.Scheme = parsedURL.Scheme
				req.URL.Host = parsedURL.Host
				return
			}
		}

		err := redisClient.SetEX(context.Background(), "sessions", ip, 60*time.Second).Err()
		if err != nil {
			log.Fatalf("Failed to set key: %v", err)
		}

		sm.Lock()
		defer sm.Unlock()
		// Default behavior (non-sticky)
		target = targets[rr]
		rr = (rr + 1) % len(targets)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		log.Printf("request sent to %s", req.URL)
	}

	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			// Add retry logic to the transport
			MaxIdleConnsPerHost:   0,
			ResponseHeaderTimeout: 10 * time.Second,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				var d net.Dialer
				var conn net.Conn
				var err error
				for i := 0; i < 3; i++ {
					conn, err = d.DialContext(ctx, network, addr)
					if err == nil {
						return conn, nil
					}
					fmt.Printf("Failed to connect to %s (attempt %d): %v\n", addr, i+1, err)
					// Wait between retries
				}
				for _, host := range targets {
					conn, err = d.DialContext(ctx, network, "localhost:8009")
					if err == nil {
						return conn, nil
					}
					fmt.Printf("Failed to connect to %s : %v\n", host, err)
					// Wait between retries
				}
				fmt.Printf("All connection attempts failed\n")
				return nil, err
			},
		},
	}
}
