package addressing

import (
	"fmt"
	"net"
	"sync"

	"github.com/munnerz/gobalancer/pkg/api"
)

var (
	// ErrAddressPoolFull means there are no addresses left in the allocators
	// network to allocate to services
	ErrAddressPoolFull = fmt.Errorf("IP address pool full")
)

// Allocator keeps track of and monitors the IPs allocated to services
// and is able to allocate new addresses from a pool to services
type Allocator struct {
	device           string
	network          net.IPNet
	allocatedIPs     []net.IP
	allocatedIPsLock sync.Mutex
}

// AllocateIP will allocate an IP address from this allocators network pool,
// or return an error if the pool is full
func (a *Allocator) AllocateIP() (*net.IPNet, error) {

	// define a function the increments an IP address
	inc := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

OuterLoop:
	for ip := a.network.IP.Mask(a.network.Mask); a.network.Contains(ip); inc(ip) {
		for _, j := range a.allocatedIPs {
			if j.Equal(ip) {
				continue OuterLoop
			}
		}
		ipn := net.IPNet{
			IP:   ip,
			Mask: a.network.Mask,
		}
		a.allocateIP(ipn.IP)

		return &ipn, nil
	}
	return nil, ErrAddressPoolFull
}

// NewAllocator initialises a new IP address allocator with the given
// device name and network
func NewAllocator(a *api.Allocator) *Allocator {
	return &Allocator{
		device:  a.Device,
		network: a.Network,
	}
}

func (a *Allocator) allocateIP(e net.IP) {
	a.allocatedIPsLock.Lock()
	defer a.allocatedIPsLock.Unlock()

	a.allocatedIPs = append(a.allocatedIPs, e)
}

func (a *Allocator) deallocateIP(e net.IP) {
	a.allocatedIPsLock.Lock()
	defer a.allocatedIPsLock.Unlock()

	eleI := -1
	for i, se := range a.allocatedIPs {
		if se.Equal(e) {
			eleI = i
			break
		}
	}

	if eleI == -1 {
		return
	}

	a.allocatedIPs = append(a.allocatedIPs[:eleI], a.allocatedIPs[eleI+1:]...)
}