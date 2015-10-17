package api

import "fmt"

func (b *Backend) String() string {
	return fmt.Sprintf("\n  Hostname: %s\n  PollInterval: %d\n  PollTimeout: %d\n ", b.Hostname, b.PollInterval, b.PollTimeout)
}
