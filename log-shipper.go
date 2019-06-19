package logshipper

import (
	"encoding/json"
	"fmt"
	"time"

	mp "github.com/vmihailenco/msgpack"
)

// Events is a channel of loggable Events
var Events = make(chan Event, 1000)

// DefaultAddr is the default endpoint for shipping/receiving logs
var DefaultAddr = "0.0.0.0:8000"

// StartUploading is intended to be called as a goroutine. Every `interval`, upload events that have not been uploaded, then delete all those that have
func StartUploading(endpoint string, mimeType string, interval time.Duration, maxBatchSize int) {
	lastUploadTime := LastUploadTime()

	if endpoint == "" {
		endpoint = "http://" + DefaultAddr
	}

	for {
		nextUploadTime := lastUploadTime.Add(interval)
		now := time.Now()
		if now.After(nextUploadTime) {
			if uploaded, err := UploadNextEvents(endpoint, mimeType, maxBatchSize); err != nil {
				fmt.Println("Error uploading events: " + err.Error())
			} else {
				if uploaded > 0 {
					fmt.Println("Uploaded")
					if err := DeleteUploadedEvents(); err != nil {
						fmt.Println("Error deleting events: " + err.Error())
					} else {
						fmt.Println("Deleted")
					}
				}
			}
			lastUploadTime = now
		} else {
			time.Sleep(nextUploadTime.Sub(now))
		}
	}
}

// UploadNextEvents gets the next Events to upload, uploads them, then marks them as uploaded
func UploadNextEvents(endpoint string, mimeType string, limit int) (int, error) {
	marshalledEvents, ks := GetEventsToUpload(limit)
	if len(ks) == 0 {
		return 0, nil
	}

	compressed := gz(marshalledEvents, "☃")
	_, err := gzPOST(endpoint, mimeType, compressed)

	if err != nil {
		return 0, err
	}

	err = MarkEventsUploaded(ks)

	if err != nil {
		return 0, err
	}

	return len(ks), nil
}

// Mp2json converts a msgpack byte slice to a JSON byte slice
func Mp2json(inBz []byte, out interface{}) ([]byte, error) {
	err := mp.Unmarshal(inBz, out)
	if err != nil {
		return nil, err
	}

	return json.Marshal(out)
}

// StartInterval calls `onTick` every `interval`
func StartInterval(interval time.Duration, onTick func(int)) chan struct{} {
	quit := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		i := 0
		for {
			select {
			case <-ticker.C:
				onTick(i)
				i++
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit
}
