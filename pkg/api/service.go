package api

import (
	"encoding/json"
	"fmt"
	"net"
)

type service struct {
	Object

	IP       *string    `json:"ip,omitempty"`
	Ports    []*PortMap `json:"ports"`
	Backends []*Backend `json:"backends"`
}

func (a *Service) MarshalJSON() ([]byte, error) {
	ip := new(string)

	if a.IP != nil {
		*ip = a.IP.String()
	}

	a2 := service{
		Object: Object{
			Name: a.Name,
		},
		IP:       ip,
		Ports:    a.Ports,
		Backends: a.Backends,
	}

	return json.Marshal(a2)
}

func (a *Service) UnmarshalJSON(b []byte) error {
	a2 := service{}

	err := json.Unmarshal(b, &a2)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &a.Object)

	if err != nil {
		return err
	}

	if a2.IP != nil {
		ip, mask, err := net.ParseCIDR(*a2.IP)

		if err != nil {
			return err
		}

		a.IP = &net.IPNet{
			IP:   ip,
			Mask: mask.Mask,
		}
	}

	a.Ports = a2.Ports
	a.Backends = a2.Backends

	return nil
}

func (a *Service) String() string {
	return fmt.Sprintf("\n Name: %s\n IP: %s\n Ports: %s\n Backends: %s\n", a.Name, a.IP, a.Ports, a.Backends)
}
