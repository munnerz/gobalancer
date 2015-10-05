package api

import (
	"encoding/json"
	"fmt"
	"net"
)

type allocator struct {
	Device  string `json:"service_net_dev"`
	Network string `json:"service_ip_range"`
}

func (a *Allocator) MarshalJSON() ([]byte, error) {
	a2 := allocator{
		Device:  a.Device,
		Network: a.Network.String(),
	}

	return json.Marshal(a2)
}

func (a *Allocator) UnmarshalJSON(b []byte) error {
	a2 := allocator{}

	err := json.Unmarshal(b, &a2)

	if err != nil {
		return err
	}

	_, network, err := net.ParseCIDR(a2.Network)

	if err != nil {
		return err
	}

	a.Device = a2.Device
	a.Network = *network

	return nil
}

func (a *Allocator) String() string {
	return fmt.Sprintf("\n Device: %s\n Network: %s\n", a.Device, a.Network.String())
}
