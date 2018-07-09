package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/passthrough-microservice/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	port := ":8018"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	secure.GET("/simple/:gw/*", handlers.SimplePassthrough)         //simple passthrough
	secure.GET("/sequenced/:gw/*", handlers.SequencedPassthrough)   //Sequence all commands (only allow one out standing request at a time)
	secure.GET("/metered/:rate/:gw/*", handlers.MeteredPassthrough) //a specific type that requires an entry in the configuration file

	secure.GET("/delayed/:delay/:gw/*/resp/*", handlers.DelayedPassthrough)

	// endpoint to handle the rmc3 responding with an error, even though the input is changing.
	//	secure.GET("/slowrmc/:delay/:address/input/:input/:output", handlers.SetRMCInput)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
