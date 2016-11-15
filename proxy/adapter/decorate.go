package adapter

// We define what a decorator to our adapter will look like
type Decorator func(Adapter) Adapter

// Decorate will decorate a adapter with a slice of passed decorators
func Decorate(c Adapter, ds ...Decorator) Adapter {
	decorated := c
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}

func ThisDecorate(c Adapter, ds []Decorator) Adapter {
	decorated := c
	for _, decorate := range ds {
		decorated = decorate(decorated)
	}
	return decorated
}
