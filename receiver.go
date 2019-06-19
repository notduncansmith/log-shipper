package logshipper

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// StartReceiving starts an HTTP server at `addr` that invokes the `eachLine` callback for each uncompressed line received
func StartReceiving(addr string, eachLine func(string, error)) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bz, err := ioutil.ReadAll(r.Body)

		if err != nil {
			fmt.Println("Unable to read body")
			w.WriteHeader(500)
			w.Write([]byte("error"))
		} else {
			lines, err := ungz(bz)

			if err != nil {
				eachLine("", err)
				return
			}

			for _, line := range lines {
				if line == "" {
					continue
				}
				eachLine(line, nil)
			}

			w.WriteHeader(200)
			w.Write([]byte("ok"))
		}
	})

	fmt.Printf("Receiving logs at %v\n", addr)
	http.ListenAndServe(addr, nil)
}
