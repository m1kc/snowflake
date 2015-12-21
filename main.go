package main

import (
	"log"
	"net/http"
	"text/template"
)

const (
	ADDR string = ":8080"
)

var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(res, req.Host)
}

func main() {
	go messageHub.run()
	go messageHub.generateTestMessages()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	log.Printf("[main] Starting server")
	log.Printf("[main] Try http://localhost%s", ADDR)
	if err := http.ListenAndServe(ADDR, nil); err != nil {
		log.Fatal(err)
	}
}
