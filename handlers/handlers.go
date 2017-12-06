package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/byuoitav/passthrough-microservice/passthrough"
	"github.com/fatih/color"
	"github.com/labstack/echo"
)

/*
SimplePassthrough just passes on the request to the address denoted in :gw, nothing more.
*/
func SimplePassthrough(context echo.Context) error {

	gw := context.Param("gw")
	prePath := "/simple/" + gw

	path := strings.Replace(context.Request().RequestURI, prePath, "", 1)

	//now we just make the request

	//http for now
	//TODO: move to https
	addr := fmt.Sprintf("http://%v%v", gw, path)

	log.Printf("Passthrough request from %v to %v", context.RealIP(), addr)

	resp, err := http.Get(addr)
	if err != nil {
		msg := fmt.Sprintf("Error with passthrough request: %v", err.Error())
		log.Printf(msg)
		return context.JSON(http.StatusInternalServerError, msg)
	}

	log.Printf("Passthrough from %v to %v done.", context.RealIP(), addr)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("Error with passthrough request: %v", err.Error())
		log.Printf(msg)
		return context.JSON(http.StatusInternalServerError, msg)
	}
	resp.Body.Close()

	return context.Blob(resp.StatusCode, resp.Header.Get("content-type"), body)
}

func SequencedPassthrough(context echo.Context) error {
	gw := context.Param("gw")
	prePath := "/sequenced/" + gw

	path := strings.Replace(context.Request().RequestURI, prePath, "", 1)
	code, rType, content, err := passthrough.SequencedPassthrough(gw, path, 0)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.Blob(code, rType, content)
}

func MeteredPassthrough(context echo.Context) error {
	gw := context.Param("gw")
	rate := context.Param("rate")
	prePath := "/metered/" + rate + "/" + gw

	i, err := strconv.Atoi(rate)
	if err != nil {
		msg := fmt.Sprintf("rate must be an integer, was %v", rate)
		log.Printf(color.HiRedString(msg))
		return context.JSON(http.StatusBadRequest, msg)
	}

	path := strings.Replace(context.Request().RequestURI, prePath, "", 1)
	code, rType, content, err := passthrough.SequencedPassthrough(gw, path, i)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.Blob(code, rType, content)
	return nil
}
