package adapter

import (
	"log"
)

// Logging will create a adapter decorator with logging concerns.
func Logging(l *log.Logger) Decorator {
	return func(c Adapter) Adapter {
		return AdapterFunc(func(r *[]byte, i int) error {
			l.Printf("log: msg len=%d\n", i)
			return c.Do(r, i)
		})
	}
}
