package api

import (
	"encoding/json"
	"fmt"
)

type servicePort struct {
	Object

	Type ServicePortType `json:"type"`
	Src  int             `json:"src"`
	Dst  int             `json:"dst"`
}

func (a *ServicePort) MarshalJSON() ([]byte, error) {
	a2 := servicePort{
		Object: Object{
			Name: a.Name,
		},
		Type: a.Type,
		Src:  a.Src,
		Dst:  a.Dst,
	}

	return json.Marshal(a2)
}

func (a *ServicePort) UnmarshalJSON(b []byte) error {
	a2 := servicePort{}

	err := json.Unmarshal(b, &a2)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &a.Object)

	if err != nil {
		return err
	}

	a.Type = a2.Type
	a.Dst = a2.Dst
	a.Src = a2.Src

	return nil
}

func (s *ServicePort) String() string {
	return fmt.Sprintf("\n  Type: %s\n  Src: %d\n  Dst: %d\n ", s.Type, s.Src, s.Dst)
}
