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

package admin

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang/glog"
	"net/http"
)

type ProxyEvent struct {
	Name    string
	Message string
}

var EventChannel []chan ProxyEvent

func init() {
	glog.V(2).Infoln("setting up the Event Channel")
}
func AddEventSubscriber() chan ProxyEvent {
	subscriber := make(chan ProxyEvent)
	EventChannel = append(EventChannel, subscriber)
	return subscriber
}

func StreamEvents(w rest.ResponseWriter, r *rest.Request) {
	cpt := 0
	//create and add a subcriber channel
	eventsChannel := AddEventSubscriber()

	for {
		cpt++
		glog.V(2).Infoln("waiting for stream channel to get event")
		select {
		case event := <-eventsChannel:
			glog.V(2).Infoln("got an Event from channel")
			w.WriteJson(&event)
			w.(http.ResponseWriter).Write([]byte("\n"))
			// Flush the buffer to client
			w.(http.Flusher).Flush()
		}

	}
}
