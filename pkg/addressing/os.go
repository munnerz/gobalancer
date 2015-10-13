package addressing

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
)

var (
	// ErrAddressBound is returned when the IP address passed to RegisterIP is already
	// bound to a network interface on the system
	ErrAddressBound  = fmt.Errorf("IP address already bound")
	ErrUnsupportedOS = fmt.Errorf("Unsupported operating system for binding IP addresses")
)

// RegisterIP will bind an IP address to this allocators network device.
// Currently this only supports OS X (darwin) and Linux
func (a *IPPool) RegisterIP(ipnet net.IPNet) error {
	if ipnet.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return ErrAddressBound
	}

	a.allocateIP(ipnet.IP)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		exec.Command("ip", "addr", "add", ipnet.String(), "dev", a.device)
	case "darwin":
		cmd = exec.Command("ifconfig", a.device, "inet", ipnet.IP.String(), "netmask", net.IP(ipnet.Mask).String(), "alias")
	case "windows":
		cmd = exec.Command("netsh", "interface", "ipv4", "add", a.device, ipnet.IP.String(), net.IP(ipnet.Mask).String())
	default:
		return ErrUnsupportedOS
	}

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("Error executing ip command: %s", err.Error())
	}
	return nil
}

// UnregisterIP will release an IP address from this allocators network device.
// Currently this only supports OS X (darwin) and Linux
func (a *IPPool) UnregisterIP(ipnet net.IPNet) error {
	if ipnet.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return nil
	}

	a.deallocateIP(ipnet.IP)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ip", "addr", "del", ipnet.String(), "dev", a.device)
	case "darwin":
		cmd = exec.Command("ifconfig", a.device, "inet", ipnet.IP.String(), "netmask", net.IP(ipnet.Mask).String(), "delete")
	case "windows":
		cmd = exec.Command("netsh", "interface", "ipv4", "delete", a.device, ipnet.IP.String(), net.IP(ipnet.Mask).String())
	default:
		return ErrUnsupportedOS
	}

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("Error executing ip command: %s", err.Error())
	}
	return nil
}
