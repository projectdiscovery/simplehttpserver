package httpserver

import "io/ioutil"

func handleUpload(file string, data []byte) error {
	return ioutil.WriteFile(file, data, 0655)
}
