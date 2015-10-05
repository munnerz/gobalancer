package addressing

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var (
	// ErrAddressBound is returned when the IP address passed to RegisterIP is already
	// bound to a network interface on the system
	ErrAddressBound = fmt.Errorf("IP address already bound")
)

func exitCode(e error) (int, error) {
	p := strings.Split(e.Error(), " ")
	return strconv.Atoi(p[len(p)-1])
}

// RegisterIP will bind an IP address to this allocators network device.
// Currently this only supports OS X (darwin) and Linux
func (a *Allocator) RegisterIP(ipnet net.IPNet) error {
	if ipnet.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return ErrAddressBound
	}

	a.allocateIP(ipnet.IP)

	cmd := exec.Command("ip", "addr", "add", ipnet.String(), "dev", a.device)

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ifconfig", a.device, "inet", ipnet.IP.String(), "netmask", ipnet.Mask.String(), "alias")
	}

	err := cmd.Run()

	if err != nil {
		c, err := exitCode(err)

		if err != nil {
			return err
		}

		switch c {
		case 1: // Invalid permissions
			return fmt.Errorf("Run loadbalancer as root to bind addresses!")
		case 2: // IP already bound
			return ErrAddressBound
		}

		return fmt.Errorf("Error executing ip command. Exit code: %d", c)
	}
	return nil
}

// UnregisterIP will release an IP address from this allocators network device.
// Currently this only supports OS X (darwin) and Linux
func (a *Allocator) UnregisterIP(ipnet net.IPNet) error {
	if ipnet.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return nil
	}

	a.deallocateIP(ipnet.IP)

	cmd := exec.Command("ip", "addr", "del", ipnet.String(), "dev", a.device)

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ifconfig", a.device, "inet", ipnet.IP.String(), "netmask", ipnet.Mask.String(), "delete")
	}

	err := cmd.Run()

	if err != nil {
		c, err := exitCode(err)

		if err != nil {
			return err
		}

		return fmt.Errorf("Error executing ip command. Exit code: %d", c)
	}
	return nil
}
