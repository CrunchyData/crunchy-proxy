package adapter

import (
	"log"
)

// Audit will create a adapter decorator with auditing concerns.
func Audit(l *log.Logger) Decorator {
	return func(c Adapter) Adapter {
		return AdapterFunc(func(r *[]byte, i int) error {
			l.Println("added to audit trail")
			return c.Do(r, i)
		})
	}
}
