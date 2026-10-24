package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	listenPort              int
	minutesTillSelfDestruct int
	numConnections          int
	connectionQueue         chan interface{}
	activeConnections       ConnectionList
	allowSameClient         bool
	hairTrigger             bool
	bombId                  string
	catchSignals            bool
)

type ConnectionList struct {
	Connections map[string]bool
	Lock        sync.Mutex
}

func init() {
	log.Println("Initializing trigger v39.37...")

	flag.IntVar(&listenPort, "listenPort", 9803, "The port on which to listen for connections")
	flag.IntVar(&minutesTillSelfDestruct, "mintoselfdestruct", 45, "The number of minutes until the system self destructs")
	flag.IntVar(&numConnections, "numConnections", 5, "The maximum number of supported connections")
	flag.BoolVar(&allowSameClient, "allowClient", true, "Allow the same client to connect multiple times")
	flag.BoolVar(&hairTrigger, "hairTrigger", false, "Detonate on client disconnect")
	flag.BoolVar(&catchSignals, "catchSignals", false, "Detonate on SIGINT/SIGTERM (the 'unknown modification' trap)")
	flag.Parse()

	//Listen for connections on this port
	tmpListenPort := os.Getenv("BOOM_LISTEN_PORT")
	if tmpListenPort != "" {
		i, err := strconv.Atoi(tmpListenPort)
		if err != nil {
			log.Fatalf("Error parsing environment variable BOOM_LISTEN_PORT (%v): %v", tmpListenPort, err)
		}
		listenPort = i
	}

	//Wait this many minutes until self destructing
	tmpMinutesTillSelfDestruct := os.Getenv("MINUTES_TILL_SELF_DESTRUCT")
	if tmpMinutesTillSelfDestruct != "" {
		i, err := strconv.Atoi(tmpMinutesTillSelfDestruct)
		if err != nil {
			log.Fatalf("Error parsing environment variable MINUTES_TILL_SELF_DESTRUCT (%v): %v", tmpMinutesTillSelfDestruct, err)
		}
		minutesTillSelfDestruct = i
	}

	//Only allow this many connections at once for safety purposes
	tmpNumConnections := os.Getenv("NUM_CONNECTIONS")
	if tmpNumConnections != "" {
		i, err := strconv.Atoi(tmpNumConnections)
		if err != nil {
			log.Fatalf("Error parsing environment variable NUM_CONNECTIONS (%v): %v", tmpNumConnections, err)
		}
		numConnections = i
	}

	if envBombId, ok := os.LookupEnv("BOMB_ID"); ok && envBombId != "" {
		bombId = " " + envBombId
	}

	//DON'T DETONATE IMMEDIATELY!!!
	if minutesTillSelfDestruct < 0 {
		log.Fatalf("Environment variable MINUTES_TILL_SELF_DESTRUCT is negative and must be positive, MINUTES_TILL_SELF_DESTRUCT = %v\n", minutesTillSelfDestruct)
	}

	if numConnections < 1 {
		log.Fatalf("Environment variable NUM_CONNECTIONS is less than 1, NUM_CONNECTIONS = %v\n", numConnections)
	}

	connectionQueue = make(chan interface{}, numConnections)
	activeConnections.Connections = make(map[string]bool)
}

func main() {

	// The real exercise ships this commented out in main() as a hint. The
	// -catchSignals flag re-enables it so you can practice the "unknown
	// modification" variant where a plain kill/Ctrl-C detonates.
	if catchSignals {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			for {
				s := <-sigChan
				handleSignal(s)
			}
		}()
	}

	log.Printf("Set up complete. Program wmd-trigger will attempt to listen on port tcp//0.0.0.0:%v and will self destruct in %v minutes.\n", listenPort, minutesTillSelfDestruct)
	go selfDestruct(minutesTillSelfDestruct)

	listenAddr := fmt.Sprintf(":%v", listenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Error starting listener for connection: %v", err)
	}
	log.Printf("Listening on tcp//%v\n", listenPort)

	for {
		log.Println("Waiting for new connection...")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}
		clientAddr := conn.RemoteAddr().String()
		clientAddrSlice := strings.Split(clientAddr, ":")
		clientIP := clientAddrSlice[0]
		log.Printf("Attempted connection from: %v\n", clientIP)
		log.Printf("Verifying client %v\n", clientIP)
		//Confirm client isn't already connected
		allowConnection := checkConnection(clientIP)
		if allowConnection {
			log.Printf("Accepted connection from %v\n", clientIP)
			//If queue is full, close the client connection
			select {
			case connectionQueue <- true:
				activeConnections.Lock.Lock()
				activeConnections.Connections[clientIP] = true
				activeConnections.Lock.Unlock()
				go connectionHandler(conn)
				log.Printf("Connection established from %v\n\tNumber of active connections: %v", clientIP, len(connectionQueue))
			default:
				log.Printf("Failed to establish connection from %v due to full queue\n", clientIP)
				conn.Close()
			}
		} else {
			log.Printf("Rejecting connection from %v\n", clientIP)
			conn.Close()
		}
	}

}

func connectionHandler(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	clientAddrSlice := strings.Split(clientAddr, ":")
	clientIP := clientAddrSlice[0]

	//Close connection and remove from active list when finished
	if hairTrigger {
		defer directDetonateBomb("HAIR")
	}
	defer conn.Close()
	defer func() {
		activeConnections.Lock.Lock()
		delete(activeConnections.Connections, clientIP)
		activeConnections.Lock.Unlock()
		<-connectionQueue
		log.Printf("\tNumber of active connections: %v", len(connectionQueue))
	}()

	log.Printf("Waiting for message from %v\n", clientIP)
	buf := make([]byte, 100)
	sizeRead, err := conn.Read(buf)
	if err != nil {
		log.Printf("Error reading message from %v: %v\n", clientIP, err)
		return
	}
	//Get detonation code from client connection
	log.Printf("Received message from %v: %v\n", clientIP, base64.StdEncoding.EncodeToString(buf[:sizeRead]))
	detCode := buf[:sizeRead]
	err = detonateBomb(string(detCode))
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Successfully ran detonation safety script")
	}

}

func selfDestruct(countdown int) {
	log.Println("Self destruct routine initiated")
	//Wait until timeout, then detonate
	for timeRemaining := countdown; timeRemaining > 0; timeRemaining-- {
		log.Printf("Self destructing in %v minutes\n", timeRemaining)
		time.Sleep(time.Minute)
	}
	log.Println("Self destructing")
	err := detonateBomb("TIME")
	if err != nil {
		log.Printf("Error detonating WMD%s: %v\n", bombId, err)
	} else {
		log.Println("Successfully ran detonation safety script")
	}

	time.Sleep(time.Minute * 4)
	log.Printf("Error detonating WMD%s: this isn't good...\n", bombId)
}

func checkConnection(ip string) bool {
	activeConnections.Lock.Lock()
	active := activeConnections.Connections[ip]
	activeConnections.Lock.Unlock()
	if active {
		log.Printf("Connection to client %v already exists\n", ip)
	} else {
		log.Printf("Connection to client %v does *not* exist\n", ip)
	}
	if allowSameClient {
		return true
	}

	return !active
}

func detonateBomb(detonationCode string) error {
	//Run bash script to check if bomb is in position and ready to detonate
	cmdString := fmt.Sprintf("./safety.sh %s", detonationCode)
	logString := base64.StdEncoding.EncodeToString([]byte(cmdString))
	log.Printf("Invoking command via system: %v\n", logString)
	cmd := exec.Command("bash", "-c", cmdString)
	output, err := cmd.CombinedOutput()
	log.Println(base64.StdEncoding.EncodeToString(output))
	if err != nil {
		log.Printf("Error running safety check: %v", err)
		return err
	}

	return nil
}

func directDetonateBomb(detonationCode string) error {
	//Run bash script to check if bomb is in position and ready to detonate
	cmdString := fmt.Sprintf("./detonate %s", detonationCode)
	logString := base64.StdEncoding.EncodeToString([]byte(cmdString))
	log.Printf("Invoking command via system: %v\n", logString)
	cmd := exec.Command("bash", "-c", cmdString)
	output, err := cmd.CombinedOutput()
	log.Println(base64.StdEncoding.EncodeToString(output))
	if err != nil {
		log.Printf("Error running detonate command: %v", err)
		return err
	}

	return nil
}

func handleSignal(signal os.Signal) {
	if signal == syscall.SIGTERM || signal == syscall.SIGINT {
		//detonate if SIGINT or SIGTERM signal recieved
		directDetonateBomb("TRIG")
	}
}
