package passthrough

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Delay i
func Delay(host, reqPath, respPath string, delay time.Duration) ([]byte, int, string, *nerr.E) {
	log.SetLevel("debug")
	log.L.Infof("Executing delayed request/response to %s.", host)

	// build the initial request
	log.L.Debugf("Sending initial request to: http://%s%s", host, reqPath)
	resp, err := http.Get(fmt.Sprintf("http://%s%s", host, reqPath))
	if err != nil {
		return []byte{}, resp.StatusCode, "", nerr.Translate(err)
	}
	defer resp.Body.Close()

	// read initial response to make sure it's good
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp.StatusCode, "", nerr.Translate(err).Addf("failed to read initial response body")
	}

	log.L.Debugf("Successfully made initial request. Response body: %s", bytes)

	// delay
	time.Sleep(delay)

	// get the response
	log.L.Debugf("Sending second request to: http://%s%s", host, respPath)
	resp, err = http.Get(fmt.Sprintf("http://%s%s", host, respPath))
	if err != nil {
		return []byte{}, resp.StatusCode, "", nerr.Translate(err).Addf("successfully executed initial request, but failed to execute second request.")
	}
	defer resp.Body.Close()

	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp.StatusCode, "", nerr.Translate(err).Addf("failed to read second response body")
	}

	log.L.Debugf("Successfully made second request. Response body: %s", bytes)
	return bytes, resp.StatusCode, resp.Header.Get("content-type"), nil
}
