package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/20ritiksingh/LoadBalancer/servers"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
)

func main() {
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

	fmt.Println("Reverse proxy server listening on port 8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
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

	director := func(req *http.Request) {

		sessionID := req.Header.Get("Session-ID")
		if sessionID != "" {
			targetURL, err := redisClient.HGet(req.Context(), "sessions", sessionID).Result()
			if err == nil && targetURL != "" {
				parsedURL, _ := url.Parse(targetURL)
				req.URL.Scheme = parsedURL.Scheme
				req.URL.Host = parsedURL.Host
				return
			}
		}
		sm.Lock()
		defer sm.Unlock()
		// Default behavior (non-sticky)
		target := targets[0]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}
