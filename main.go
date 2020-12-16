package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var debugMode bool

func main() {
	listenAddress := flag.String("listen-address", "", "the address on which to listen")
	listenPort := flag.Int("listen-port", 8080, "the port on which to listen")
	debugModeFlag := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	debugMode = *debugModeFlag

	if !debugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	peopleHrCalendarUrl, gotUrl := os.LookupEnv("PEOPLEHR_CALENDAR_URL")
	if !gotUrl {
		log.Fatalf("Please specify PEOPLEHR_CALENDAR_URL in the environment!\n")
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

	listenAddressAndPort := fmt.Sprintf("%s:%d", *listenAddress, *listenPort)
	log.Printf("API server listening on %s", listenAddressAndPort)

	err := http.ListenAndServe(listenAddressAndPort, r)
	if err != nil {
		log.Fatalf("Fatal error:\n Failed to create HTTP server\n  %s", err)
	}
}
