package main

import (
	"flag"
	"log"
	"net/http"
	"html/template"
)

var addr = flag.String("addr", ":8080", "HTTP address to listen on")
var temp = template.Must(template.ParseFiles("growup.html"))
var msgHistory = &history{messages: make([][]byte, 0), writer: make(chan []byte), reader: make(chan chan []byte)}

type TemplateData struct {
	Host string
	History []string
}

func serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	msgs := msgHistory.dump()
	tempMsgs := make([]string, len(msgs))
	for _, msg := range msgs {
		tempMsgs = append(tempMsgs, string(msg))
	}
	err := temp.Execute(w, TemplateData{Host: r.Host, History: tempMsgs})
	if err != nil {
		log.Println("temp.Execute: ", err)
		return
	}
}

func main() {
	flag.Parse()
	go wsHub.run()
	go msgHistory.listen()
	http.HandleFunc("/", serve)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
