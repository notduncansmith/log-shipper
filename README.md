# log-shipper

log-shipper is a Go library that ships logs. Event structs are serialized (using msgpack by default) and stored in a local bbolt database, then uploaded to a given URL at a given interval and deleted locally.

## Install

`go get github.com/notduncansmith/log-shipper`

## Usage

```go
package main

import (
	"fmt"
	"strconv"
	"time"

	ls "github.com/notduncansmith/log-shipper"
)

func main() {
	// First, we must establish a connection to storage (local bbolt db)
	ls.InitStorage()

	// Next, to have something to log, we'll emit a tick event every second
	go ls.StartInterval(1*time.Second, func(i int) {
		ls.Events <- NewMyEvent("session-1", "tick "+strconv.Itoa(i))
	})

	// Then, we'll upload our logs every 3 seconds, with at most 5000 events per upload
	uploadURL := "http://" + ls.DefaultAddr // http://0.0.0.0:8000
	uploadInterval := time.Duration(3 * time.Second)
	go ls.StartUploading(uploadURL, "application/msgpack", uploadInterval, 5000)

	// This would normally run on a server at `uploadURL`, but we'll do it here for demonstration purposes
	ls.StartReceiving(ls.DefaultAddr, func(line string, err error) {
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			return
		}

		jbz, err := ls.Mp2json([]byte(line))
		if err != nil {
			fmt.Println("ERR: " + err.Error())
		} else {
			fmt.Println("RCV: " + string(jbz))
		}
	})
}

// MyEvent is a demonstration of a custom event type
type MyEvent struct {
	*ls.BaseEvent
	SessionID string `msgpack:"sessionId" json:"sessionId"`
}

// GetBytes returns the msgpack serialization of `me`
func (me MyEvent) GetBytes() []byte {
	return me.BaseEvent.GetBytesOf(me)
}

// NewMyEvent Creates a new MyEvent
func NewMyEvent(sessionID string, detail string) MyEvent {
	ev := ls.NewBaseEvent(detail)
	return MyEvent{&ev, sessionID}
}

```

**This project is graciously sponsored by Dubsado ❤️**

[![Dubsado CRM](https://global-uploads.webflow.com/5bd3a12688389fdba3a24e77/5bd3a12688389f0bc7a24ea8_dubsado-logo.png)](https://dubsado.com)

## License

MIT License

Copyright © 2019 Duncan Smith
