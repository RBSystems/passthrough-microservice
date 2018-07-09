package passthrough

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
)

var Routines = sync.Map{}

//Sequenced passthrough will take a request, and make sure it is passed in serial to all other requests to the same gateway
//returns the status code, type, and body of the response recieved.
func SequencedPassthrough(gw, path string, delayInterval int) (int, string, []byte, error) {

	respChan := make(chan response)

	req := request{
		GW:       gw,
		Path:     path,
		Interval: delayInterval,
		RespChan: respChan,
	}

	//check to see if if the gw we're interested in already has a routine associated with it

	if val, ok := Routines.Load(gw); ok {
		//we pass the request down the channel
		val.(channelWrapper).Channel <- req
	} else {
		//make a new Channel

		wrap := channelWrapper{make(chan request, 5000)}
		temp, there := Routines.LoadOrStore(gw, wrap)

		actualWrap := temp.(channelWrapper)
		if !there { //we need to start a routine
			log.Printf(color.CyanString("Starting a routine to serialize connections to %v", gw))
			go SequenceWorker(actualWrap.Channel, gw)
		}
		actualWrap.Channel <- req
	}

	//wait for the response, with a longer timeout (25 seconds)
	timer := time.NewTimer(25 * time.Second)

	select {
	case <-timer.C:
		//timeout
		log.Printf(color.HiCyanString("[Passthrough Helper] Timeout, returning"))
		return 0, "", []byte{}, errors.New("gateway timed out")
	case resp, ok := <-respChan:
		if !ok {
			return 0, "", []byte{}, errors.New("response channel from the worker was closed")
		}
		return resp.RespCode, resp.ContentType, resp.Content, resp.Error
	}

	return 0, "", []byte{}, errors.New("Internal Error: select statment escaped")
}

func SequenceWorker(incoming chan request, gw string) {
	wait := 0
	timer := time.NewTimer(24 * time.Hour)
	for {
		if wait != 0 {
			log.Printf(color.BlueString("[Worker-%v] Metered connection, waiting for %v milliseconds", gw, wait))
		}
		time.Sleep(time.Duration(wait) * time.Millisecond)

		select {
		case v, ok := <-incoming:
			if !ok {
				log.Printf(color.BlueString("[Worker-%v] Channel closed, returning.", gw))
			}
			log.Printf(color.BlueString("[Worker-%v] forwarding request received.", gw))
			if v.GW != gw {
				//gw doesn't match

				msg := fmt.Sprintf("[Worker-%v] gateway doesn't correspond to this worker. discarding. ", gw)
				log.Printf(color.HiBlueString(msg))
				v.RespChan <- response{0, "", []byte{}, errors.New(msg)}
				continue
			}
			wait = v.Interval

			addr := fmt.Sprintf("http://%v%v", gw, v.Path)
			log.Printf(color.BlueString("[Worker-%v] forwarding request to %v.", gw, addr))

			//we need a timeout here.
			timeout := time.Duration(15 * time.Second)
			client := http.Client{
				Timeout: timeout,
			}

			resp, err := client.Get(addr)
			if err != nil {
				msg := fmt.Sprintf("[Worker-%v] Error with passthrough request: %v", gw, err.Error())
				log.Printf(color.HiBlueString(msg))
				v.RespChan <- response{0, "", []byte{}, errors.New(msg)}
				continue
			}
			defer resp.Body.Close()

			log.Printf(color.BlueString("[Worker-%v] Done.", gw))

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				msg := fmt.Sprintf("[Worker-%v] Error with passthrough request: %v", gw, err.Error())
				log.Printf(color.HiBlueString(msg))
				v.RespChan <- response{0, "", []byte{}, errors.New(msg)}
				continue
			}
			resp.Body.Close()

			v.RespChan <- response{resp.StatusCode, resp.Header.Get("content-type"), body, nil}
		case <-timer.C:
			log.Printf(color.HiGreenString("[Worker-%v] timer expired, closing routine.", gw))
			Routines.Delete(gw)
			return
		}
	}
}
