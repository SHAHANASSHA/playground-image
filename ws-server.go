package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow any origin for demo purposes
	},
}

type ResizeMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type PingMessage struct {
	Type string `json:"type"`
}

type TestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type PortID struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

var vm_map = map[string]string{
	"vm1": "172.16.0.2",
	"vm2": "172.16.0.3",
	"vm3": "172.16.0.4",
	"vm4": "172.16.0.5",
	"vm5": "172.16.0.6",
}

var target_cache = map[string]string{
	"vm1p80": "172.16.0.2:80",
}

var proxy_cache = map[string]*httputil.ReverseProxy{}

// target_cache["vm1p80"] = "172.16.0.2:80"

func getProxy(id string) (*httputil.ReverseProxy, bool) {
	target, ok := target_cache[id]
	if !ok || target == "" {
		return nil, false
	}

	cachedProxy, POk := proxy_cache[id]
	if POk {
		return cachedProxy, true
	}

	targetURL, err := url.Parse("http://" + target)
	if err != nil {
		log.Fatal("Error parsing target URL: ", err)
		return nil, false
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-For", "")
		req.Header.Set("X-Real-IP", "")
		req.Host = targetURL.Host
	}
	proxy_cache[id] = reverseProxy

	return reverseProxy, true
}

func handlePublicPort(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var entry PortID
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		target_cache[entry.Key] = entry.Target
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Port opened"))
		return
	}
	if r.Method == "DELETE" {
		vars := mux.Vars(r)
		id := vars["id"]
		delete(target_cache, id)
		delete(proxy_cache, id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Port closed"))
		return
	}
	if r.Method == "GET" {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if ok {
			data := PortID{
				Key:    id,
				Target: target_cache[id],
			}
			json.NewEncoder(w).Encode(data)
			return
		}
		json.NewEncoder(w).Encode(target_cache)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func handleWebsocket(w http.ResponseWriter, r *http.Request, ip string, VMID string) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// Read SSH private key
	privateKeyBytes, err := ioutil.ReadFile("/root/firecracker/keys/ubuntu-24.04.id_rsa")
	if err != nil {
		log.Printf("Failed to read private key: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to read private key: %v", err)))
		return
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Printf("Failed to parse private key: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to parse private key: %v", err)))
		return
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH server
	sshConn, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		log.Printf("SSH connection error: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SSH Connection error: %v", err)))
		return
	}
	defer sshConn.Close()

	log.Println("SSH Connection established.")

	// Create SSH session
	session, err := sshConn.NewSession()
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating session: %v", err)))
		return
	}
	defer session.Close()

	// Set up SSH terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Printf("Failed to request PTY: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error requesting PTY: %v", err)))
		return
	}

	// Set up stdin and stdout pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Printf("Failed to create stdin pipe: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating stdin pipe: %v", err)))
		return
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating stdout pipe: %v", err)))
		return
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Printf("Failed to create stderr pipe: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating stderr pipe: %v", err)))
		return
	}

	// Start shell
	if err := session.Shell(); err != nil {
		log.Printf("Failed to start shell: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error starting shell: %v", err)))
		return
	}

	// Handle messages from WebSocket to SSH
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				session.Close()
				return
			}

			var pingMessage PingMessage
			if err := json.Unmarshal(msg, &pingMessage); err == nil && pingMessage.Type == "ping" {
				if err := ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`)); err != nil {
					log.Printf("Failed to resize PTY: %v", err)
				}
				continue
			}

			var resizeMsg ResizeMessage
			if err := json.Unmarshal(msg, &resizeMsg); err == nil && resizeMsg.Type == "resize" {
				// Adjust PTY size
				if err := session.WindowChange(resizeMsg.Rows, resizeMsg.Cols); err != nil {
					log.Printf("Failed to resize PTY: %v", err)
				}
				continue
			}

			if _, err := stdin.Write(msg); err != nil {
				log.Printf("SSH write error: %v", err)
				return
			}
		}
	}()

	// Handle stdout from SSH to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if err != nil {
				log.Printf("SSH stdout read error: %v", err)
				ws.Close()
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	// Handle stderr from SSH to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				log.Printf("SSH stderr read error: %v", err)
				return
			}
			message := fmt.Sprintf("STDERR: %s", buffer[:n])
			if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	// Wait for SSH session to end
	if err := session.Wait(); err != nil {
		log.Printf("SSH session ended with: %v", err)
	}
}

func waitForSSH(host string, port int, timeout, retryInterval time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, retryInterval)
		if err == nil {
			conn.Close()
			return nil // SSH service is ready
		}
		time.Sleep(retryInterval)
	}
	return fmt.Errorf("SSH service not available on %s:%d after %s", host, port, timeout)
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	timeout := 10 * time.Second
	retryInterval := 500 * time.Millisecond

	vms := []struct {
		ip   string
		name string
	}{
		{"172.16.0.2", "vm1"},
		{"172.16.0.3", "vm2"},
	}

	for _, vm := range vms {
		wg.Add(1)
		go func(vmIP, vmName string) {
			defer wg.Done()
			log.Printf("Checking SSH readiness for %s (%s)...", vmName, vmIP)
			err := waitForSSH(vmIP, 22, timeout, retryInterval)
			if err != nil {
				errChan <- fmt.Errorf("%s: %v", vmName, err)
			} else {
				log.Printf("SSH is ready on %s (%s)", vmName, vmIP)
			}
		}(vm.ip, vm.name)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		http.Error(w, "Some VMs are not ready:\n"+fmt.Sprintf("%v", errors), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Both VMs are ready."))
}

func isRequestSuccessful(url string) bool {
	data := map[string]string{
		"username": "johndoe",
		"password": "secret",
	}
	jsonData, err := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func examinerCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vm := vars["vm"]
	test := vars["test"]
	query := r.URL.Query().Get("args")

	ip, exist := vm_map[vm]

	w.Header().Set("Content-Type", "application/json")
	if !exist {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TestResponse{
			Success: false,
			Error:   "VM not found",
			Message: "Test failed",
		})
		return
	}
	success := isRequestSuccessful("http://" + ip + ":56678/examiner/test/" + test + "?args=" + query)
	w.WriteHeader(http.StatusOK)
	if success {
		json.NewEncoder(w).Encode(TestResponse{
			Success: true,
			Message: "Success",
		})
		return
	}
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Message: "Test failed",
	})
}

func main() {
	go func() {
		router := mux.NewRouter()

		router.HandleFunc("/examiner/test/{vm}/{test}", func(w http.ResponseWriter, r *http.Request) {
			examinerCheck(w, r)
		})
		router.HandleFunc("/vm1", func(w http.ResponseWriter, r *http.Request) {
			handleWebsocket(w, r, "172.16.0.2", "vm1")
		})
		router.HandleFunc("/vm2", func(w http.ResponseWriter, r *http.Request) {
			handleWebsocket(w, r, "172.16.0.3", "vm2")
		})
		router.HandleFunc("/wait-for-vms", func(w http.ResponseWriter, r *http.Request) {
			handleCheck(w, r)
		})
		router.HandleFunc("/open-close-port", func(w http.ResponseWriter, r *http.Request) {
			handlePublicPort(w, r)
		})
		router.HandleFunc("/open-close-port/{id}", func(w http.ResponseWriter, r *http.Request) {
			handlePublicPort(w, r)
		})

		// http.Handle("/", router)

		log.Println("WebSocket server listening on port 8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		subdomain := strings.Split(r.Host, ".")[0]
		split := strings.Split(subdomain, "-")
		id := split[len(split)-1]

		reverseProxy, ok := getProxy(id)
		if ok {
			reverseProxy.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad gateway"))
	})

	log.Println("Reverse proxy server listening on port 80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal("ListenAndServe (80): ", err)
	}
}
