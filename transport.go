package logshipper

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
)

var client = http.Client{}

const lineSeparator = "☃"

// gzPOST sends a POST request with the specified MIME type and the gzip content-encoding header (does not actually perform gzipping)
func gzPOST(url string, mimeType string, body *bytes.Buffer) ([]byte, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("Content-Encoding", "gzip")
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// gz takes a slice of byte slices and returns the gzip-compressed bytes
func gz(slices [][]byte, separator string) *bytes.Buffer {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()

	w.Write(bytes.Join(slices, []byte(lineSeparator)))

	return &b
}

// ungz decompresses a gzip-compressed byte slice
func ungz(zipped []byte) ([]string, error) {
	zr, err := gzip.NewReader(bytes.NewBuffer(zipped))
	defer zr.Close()

	if err != nil {
		return nil, err
	}
	bz, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bz), lineSeparator)
	return lines, nil
}
