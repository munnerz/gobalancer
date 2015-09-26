package logging

import (
	"fmt"
)

func Log(n string, f func(string, ...interface{}), s string, sf ...interface{}) {
	f("%s: %v", n, fmt.Sprintf(s, sf...))
}
