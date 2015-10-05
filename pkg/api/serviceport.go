package api

import "fmt"

func (s *ServicePort) String() string {
	return fmt.Sprintf("\n  Type: %s\n  Src: %d\n  Dst: %d\n ", s.Type, s.Src, s.Dst)
}
