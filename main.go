package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

var debugMode bool

func main() {
	listenAddressPtr := flag.String("listen-address", "", "the address on which to listen")
	listenPortPtr := flag.Int("listen-port", 0, "the port on which to listen")
	debugModePtr := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	listenAddress := *listenAddressPtr
	listenPort := *listenPortPtr
	debugMode := *debugModePtr

	if !debugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	if listenPort == 0 {
		// Grab the listen port from the environment if it exists.
		listenPortStr, gotPort := os.LookupEnv("NOMAD_PORT_http")
		if gotPort {
			var err error
			listenPort, err = strconv.Atoi(listenPortStr)
			if err != nil {
				log.Fatal("Failed to convert port from environment to integer!")
			}
		}
	}

	peopleHrCalendarUrl, gotUrl := os.LookupEnv("PEOPLEHR_CALENDAR_URL")
	if !gotUrl {
		log.Fatal("Please specify PEOPLEHR_CALENDAR_URL in the environment!")
	}
	peopleHr := NewPeopleHr(peopleHrCalendarUrl)

	r := gin.Default()
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		res, err := peopleHr.Today()
		if err != nil {
			c.JSON(500, gin.H{"error": err})
		} else {
			c.JSON(200, res)
		}
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddress, listenPort))
	if err != nil {
		panic(err)
	}

	log.Printf("API server listening on %s:%d\n", listenAddress, listener.Addr().(*net.TCPAddr).Port)
	panic(http.Serve(listener, r))
}
