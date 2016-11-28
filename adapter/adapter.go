package adapter

// Our adapter is defined as something that does a http request and gets a response and error
type Adapter interface {
	Do(*[]byte, int) error
}

// Singature of DoFunc
type AdapterFunc func(*[]byte, int) error

// Add method to DoFunc type to satisfy Adapter interface
func (f AdapterFunc) Do(r *[]byte, i int) error {
	return f(r, i)
}
