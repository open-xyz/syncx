package balancing

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Endpoint represents a backend server for load balancing
type Endpoint struct {
	URL        *url.URL
	IsAlive    bool
	ReverseProxy *httputil.ReverseProxy
	LastChecked time.Time
}

// Balancer represents a simple load balancer
type Balancer struct {
	Endpoints    []*Endpoint
	Current      int
	Mutex        sync.RWMutex
	CheckInterval time.Duration
}

// NewBalancer creates a new load balancer with the given endpoint URLs
func NewBalancer(endpoints []string) (*Balancer, error) {
	var serverEndpoints []*Endpoint
	
	for _, endpoint := range endpoints {
		url, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		
		proxy := httputil.NewSingleHostReverseProxy(url)
		serverEndpoint := &Endpoint{
			URL:          url,
			IsAlive:      true,
			ReverseProxy: proxy,
			LastChecked:  time.Now(),
		}
		
		serverEndpoints = append(serverEndpoints, serverEndpoint)
	}
	
	balancer := &Balancer{
		Endpoints:     serverEndpoints,
		Current:       0,
		CheckInterval: 60 * time.Second,
	}
	
	// Start health checking
	go balancer.healthCheck()
	
	return balancer, nil
}

// NextEndpoint returns the next available endpoint using round-robin
func (b *Balancer) NextEndpoint() *Endpoint {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	
	// Get initial position
	pos := b.Current
	
	// Find the next alive server
	for {
		pos = (pos + 1) % len(b.Endpoints)
		if b.Endpoints[pos].IsAlive {
			b.Current = pos
			return b.Endpoints[pos]
		}
		if pos == b.Current {
			// We've gone full circle and found no alive servers
			log.Println("Warning: No alive endpoints found")
			// Reset all endpoints to alive for retry
			for i := range b.Endpoints {
				b.Endpoints[i].IsAlive = true
			}
			return b.Endpoints[pos]
		}
	}
}

// ServeHTTP implements the http.Handler interface for the load balancer
func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endpoint := b.NextEndpoint()
	log.Printf("Routing request for %s to %s", r.URL.Path, endpoint.URL.String())
	endpoint.ReverseProxy.ServeHTTP(w, r)
}

// healthCheck periodically checks if endpoints are alive
func (b *Balancer) healthCheck() {
	ticker := time.NewTicker(b.CheckInterval)
	for {
		<-ticker.C
		b.checkEndpoints()
	}
}

// checkEndpoints checks the health of all endpoints
func (b *Balancer) checkEndpoints() {
	for i, endpoint := range b.Endpoints {
		status := "up"
		alive := isAlive(endpoint.URL)
		
		b.Mutex.Lock()
		b.Endpoints[i].IsAlive = alive
		b.Endpoints[i].LastChecked = time.Now()
		b.Mutex.Unlock()
		
		if !alive {
			status = "down"
		}
		log.Printf("Endpoint %s status: %s", endpoint.URL.String(), status)
	}
}

// isAlive checks if an endpoint is alive by sending a HEAD request
func isAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Head(u.String())
	if err != nil {
		log.Printf("Error checking endpoint %s: %v", u.String(), err)
		return false
	}
	
	if resp.StatusCode >= 500 {
		log.Printf("Endpoint %s returned status: %d", u.String(), resp.StatusCode)
		return false
	}
	
	return true
} 