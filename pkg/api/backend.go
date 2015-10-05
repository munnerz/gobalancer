package api

import "fmt"

func (b *Backend) String() string {
	return fmt.Sprintf("\n  IP: %s\n  PollInterval: %d\n  PollTimeout: %d\n ", b.IP, b.PollInterval, b.PollTimeout)
}
