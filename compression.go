package mnistdemo

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(data []byte) []byte {
	var dest bytes.Buffer
	source := bytes.NewBuffer(data)
	zip := gzip.NewWriter(&dest)
	_, err := io.Copy(zip, source)
	if err != nil {
		panic(err)
	}
	err = zip.Close()
	if err != nil {
		panic(err)
	}
	return dest.Bytes()
}

func decompress(data []byte) ([]byte, error) {
	var dest bytes.Buffer
	source := bytes.NewBuffer(data)
	zip, err := gzip.NewReader(source)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&dest, zip)
	if err != nil {
		return nil, err
	}
	return dest.Bytes(), nil
}
