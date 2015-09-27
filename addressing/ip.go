package addressing

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func exitCode(e error) (int, error) {
	p := strings.Split(e.Error(), " ")
	return strconv.Atoi(p[len(p)-1])
}

func RegisterIP(ip, subnet net.IP, device string) error {
	cmd := exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%s", ip, subnet), "dev", device)

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ifconfig", device, "inet", ip.String(), "netmask", subnet.String(), "alias")
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
			return nil
		}

		return fmt.Errorf("Error executing ip command. Exit code: %d", c)
	}
	return nil
}

func UnregisterIP(ip, subnet net.IP, device string) error {
	cmd := exec.Command("ip", "addr", "del", fmt.Sprintf("%s/%s", ip, subnet), "dev", device)

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ifconfig", device, "inet", ip.String(), "netmask", subnet.String(), "delete")
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
