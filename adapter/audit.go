/*
Copyright 2017 Crunchy Data Solutions, Inc.
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

import (
	"fmt"
	"log"
	"os"
	"time"
)

var file *os.File

func createFile(metadata map[string]interface{}, l *log.Logger) {

	var err error

	var filepath = metadata["filepath"].(string)
	if filepath == "" {
		filepath = "/tmp/audit.log"
	}
	l.Println("audit adapter using filepath of " + filepath)

	file, err = os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err.Error())
	}
}

// Audit will create a adapter decorator with auditing concerns.
func Audit(metadata map[string]interface{}, l *log.Logger) Decorator {
	return func(c Adapter) Adapter {
		return AdapterFunc(func(r []byte, i int) error {
			now := time.Now()
			if file == nil {
				createFile(metadata, l)
			}
			file.WriteString(fmt.Sprintf("%s msg len=%d\n", now.String(), i))
			file.Sync()
			return c.Do(r, i)
		})
	}
}
