package main

import (
	"log"
	"time"

	"github.com/cryptk/spring-startup-graphviz/internal/grapher"
)

func main() {

	// TODO: get this from a config param
	minDuration, err := time.ParseDuration("0.5s")
	if err != nil {
		log.Fatal(err)
	}

	graph, err := grapher.New(minDuration)
	if err != nil {
		log.Fatal(err)
	}
	defer graph.Close()

	log.Println("Fetching actuator startup information")
	// TODO: get this URL from a config param
	if err := graph.ParseURL("http://localhost:8081/actuator/startup"); err != nil {
		log.Fatal(err)
	}

	log.Println("Generating graphviz data")
	if err := graph.Generate(); err != nil {
		log.Fatal(err)
	}

	log.Println("Rendering graph to file")
	// TODO: get this file destination from a config param
	if err := graph.RenderSVGFile("/tmp/spring-boot-startup.svg"); err != nil {
		log.Fatal(err)
	}

	log.Println("All done!")

}
