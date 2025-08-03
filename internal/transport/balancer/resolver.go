package balancer

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/appnet-org/arpc/internal/transport/balancer/random"
	"github.com/appnet-org/arpc/internal/transport/balancer/types"
)

// Resolver handles DNS resolution and load balancing
type Resolver struct {
	balancer types.Balancer
}

// NewResolver creates a new resolver with the specified balancer
func NewResolver(balancer types.Balancer) *Resolver {
	return &Resolver{
		balancer: balancer,
	}
}

// ResolveUDPTarget resolves a UDP address string that may be an IP, FQDN, or empty.
// If it's empty or ":port", it binds to 0.0.0.0:<port>. For FQDNs, it uses the configured balancer
// to select an IP from the resolved addresses.
func (r *Resolver) ResolveUDPTarget(addr string) (*net.UDPAddr, error) {
	if addr == "" {
		return &net.UDPAddr{IP: net.IPv4zero, Port: 0}, nil
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// Handle addr like ":11000"
		if after, ok := strings.CutPrefix(addr, ":"); ok {
			portStr = after
			host = ""
		} else {
			return nil, fmt.Errorf("invalid addr %q: %w", addr, err)
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port in %q: %w", addr, err)
	}

	if host == "" {
		return &net.UDPAddr{IP: net.IPv4zero, Port: port}, nil
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return &net.UDPAddr{IP: ip, Port: port}, nil
	}

	// FQDN case: resolve all IPs and use balancer
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("DNS lookup failed for %q: %w", host, err)
	}

	// Log all resolved IPs
	log.Printf("DNS lookup for %s returned IPs:", host)
	for i, resolvedIP := range ips {
		log.Printf("  [%d] %s", i, resolvedIP.String())
	}

	// Use the balancer to pick an IP
	chosen := r.balancer.Pick(host, ips)
	if chosen == nil {
		return nil, fmt.Errorf("balancer failed to select an IP for %q", host)
	}

	log.Printf("Balancer '%s' selected %s → %s:%d", r.balancer.Name(), addr, chosen, port)
	return &net.UDPAddr{IP: chosen, Port: port}, nil
}

// DefaultResolver creates a resolver with a random balancer (for backward compatibility)
func DefaultResolver() *Resolver {
	return NewResolver(random.NewRandomBalancer())
}
