/*
Copyright 2016 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
