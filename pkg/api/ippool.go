package api

import (
	"encoding/json"
	"fmt"
	"net"
)

type ipPool struct {
	Name    string  `json:"name"`
	Device  string  `json:"service_net_dev"`
	Network string  `json:"service_net"`
	Range   IPRange `json:"service_net_range"`
}

func (a *IPPool) MarshalJSON() ([]byte, error) {
	a2 := ipPool{
		Name:    a.Name,
		Device:  a.Device,
		Network: a.Network.String(),
		Range:   a.Range,
	}

	return json.Marshal(a2)
}

func (a *IPPool) UnmarshalJSON(b []byte) error {
	a2 := ipPool{}

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
	a.Range = a2.Range

	err = json.Unmarshal(b, &a.Object)

	if err != nil {
		return err
	}

	return nil
}

func (a *IPPool) String() string {
	return fmt.Sprintf("\n Device: %s\n Network: %s\n, Range: %s\n", a.Device, a.Network.String(), a.Range)
}
