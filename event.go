package logshipper

import (
	"fmt"
	"time"

	mp "github.com/vmihailenco/msgpack"
)

// Event describes the minimum structure of telemetry logs
type Event interface {
	GetTimestamp() int64
	GetDetail() string
	GetBytes() []byte
}

// BaseEvent implements event
type BaseEvent struct {
	Timestamp int64  `msgpack:"timestamp" json:"timestamp"`
	Detail    string `msgpack:"detail" json:"detail"`
}

// NewBaseEvent generates a new event with specified details for the customer
func NewBaseEvent(detail string) BaseEvent {
	ts := time.Now().Unix()
	return BaseEvent{ts, detail}
}

// GetTimestamp returns the Unix timestamp
func (b BaseEvent) GetTimestamp() int64 {
	return b.Timestamp
}

// GetDetail returns the event detail string
func (b BaseEvent) GetDetail() string {
	return b.Detail
}

// GetBytes serializes the event using msgpack
func (b BaseEvent) GetBytes() []byte {
	bz, err := mp.Marshal(b)

	if err != nil {
		fmt.Printf("Unable to encode event as MsgPack (%v)", b)
		return []byte{}
	}

	return bz
}

// GetBytesOf is a convenience method to msgpack serialize a given Event, or return an empty byte slice if there was an error
func (b BaseEvent) GetBytesOf(e Event) []byte {
	bz, err := mp.Marshal(e)

	if err != nil {
		fmt.Printf("Unable to encode event as MsgPack (%v)", e)
		return []byte{}
	}

	return bz
}
