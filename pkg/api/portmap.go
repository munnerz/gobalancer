package api

import (
	"encoding/json"
	"fmt"
)

type portMap struct {
	Object

	Protocol PortMapProtocol `json:"type"`
	Src      int             `json:"src"`
	Dst      int             `json:"dst"`
}

func (a *PortMap) MarshalJSON() ([]byte, error) {
	a2 := portMap{
		Object: Object{
			Name: a.Name,
		},
		Protocol: a.Protocol,
		Src:      a.Src,
		Dst:      a.Dst,
	}

	return json.Marshal(a2)
}

func (a *PortMap) UnmarshalJSON(b []byte) error {
	a2 := portMap{}

	err := json.Unmarshal(b, &a2)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &a.Object)

	if err != nil {
		return err
	}

	a.Protocol = a2.Protocol
	a.Dst = a2.Dst
	a.Src = a2.Src

	return nil
}

func (s *PortMap) String() string {
	return fmt.Sprintf("\n  Protocol: %s\n  Src: %d\n  Dst: %d\n ", s.Protocol, s.Src, s.Dst)
}
