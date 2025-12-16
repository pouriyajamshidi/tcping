// Package dns handles all hostname resolution logic
package dns

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/option"
)

var (
	ErrNoIPv4Address = errors.New("no ipv4 address found")
	ErrNoIPv6Address = errors.New("no ipv6 address found")
	ErrNoIPAddresses = errors.New("no ip addresses")
	ErrResolve       = errors.New("resolve hostname")
)

// Resolver handles hostname resolution with configurable options
type Resolver struct {
	timeout time.Duration
	useIPv4 bool
	useIPv6 bool
}

type ResolverOption = option.Option[Resolver]

// WithTimeout sets the DNS resolution timeout
func WithTimeout(timeout time.Duration) ResolverOption {
	return func(r *Resolver) {
		r.timeout = timeout
	}
}

// WithIPv4Only configures the resolver to only return IPv4 addresses
func WithIPv4Only() ResolverOption {
	return func(r *Resolver) {
		r.useIPv4 = true
		r.useIPv6 = false
	}
}

// WithIPv6Only configures the resolver to only return IPv6 addresses
func WithIPv6Only() ResolverOption {
	return func(r *Resolver) {
		r.useIPv4 = false
		r.useIPv6 = true
	}
}

const (
	defaultTimeout = 2 * time.Second
	ipv4OrIPv6     = "ip" // allows LookupNetIP to use both IPv4 and IPv6
)

// NewResolver creates a new DNS resolver with optional configuration
func NewResolver(opts ...ResolverOption) *Resolver {
	r := &Resolver{
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// ResolveHostname resolves a hostname to an IP address respecting the context deadline
func (r *Resolver) ResolveHostname(ctx context.Context, hostname string) (netip.Addr, error) {
	ip, err := netip.ParseAddr(hostname)
	if err == nil {
		return ip, nil
	}

	lctx := ctx
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		lctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	ipAddrs, err := net.DefaultResolver.LookupNetIP(lctx, ipv4OrIPv6, hostname)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("%w: %s: %w", ErrResolve, hostname, err)
	}

	var filtered []netip.Addr
	switch {
	case r.useIPv4:
		filtered = filterIPv4(ipAddrs)
		if len(filtered) == 0 {
			return netip.Addr{}, fmt.Errorf("%w: %s", ErrNoIPv4Address, hostname)
		}
	case r.useIPv6:
		filtered = filterIPv6(ipAddrs)
		if len(filtered) == 0 {
			return netip.Addr{}, fmt.Errorf("%w: %s", ErrNoIPv6Address, hostname)
		}
	default:
		filtered = unmapAddresses(ipAddrs)
	}

	return selectRandomIP(filtered)
}

// ResolveHostname is a package-level convenience function that uses default settings
func ResolveHostname(ctx context.Context, hostname string, useIPv4, useIPv6 bool) (netip.Addr, error) {
	var opts []ResolverOption
	if useIPv4 {
		opts = append(opts, WithIPv4Only())
	} else if useIPv6 {
		opts = append(opts, WithIPv6Only())
	}
	r := NewResolver(opts...)
	return r.ResolveHostname(ctx, hostname)
}

func selectRandomIP(ipAddrs []netip.Addr) (netip.Addr, error) {
	if len(ipAddrs) == 0 {
		return netip.Addr{}, ErrNoIPAddresses
	}
	return ipAddrs[rand.Intn(len(ipAddrs))], nil
}

func filterIPv4(ipAddrs []netip.Addr) []netip.Addr {
	var ipList []netip.Addr
	for _, ip := range ipAddrs {
		// static builds (CGO=0) return IPv4-mapped IPv6 addresses
		if ip.Is4() || ip.Is4In6() {
			ipList = append(ipList, ip.Unmap())
		}
	}
	return ipList
}

func filterIPv6(ipAddrs []netip.Addr) []netip.Addr {
	var ipList []netip.Addr
	for _, ip := range ipAddrs {
		if ip.Is6() && !ip.Is4In6() {
			ipList = append(ipList, ip)
		}
	}
	return ipList
}

func unmapAddresses(ipAddrs []netip.Addr) []netip.Addr {
	ipList := make([]netip.Addr, len(ipAddrs))
	for i, ip := range ipAddrs {
		ipList[i] = ip.Unmap()
	}
	return ipList
}
