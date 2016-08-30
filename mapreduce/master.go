package mapreduce

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

const (
	IDLE_WORKER_BUFFER = 10
)

type Master struct {
	// Network data
	address   string
	rpcServer *rpc.Server
	listener  net.Listener

	// Workers handling
	workersMutex   sync.Mutex
	idleWorkerChan chan *RemoteWorker

	workers      []*RemoteWorker
	totalWorkers int // Used to generate unique ids for new workers

	// Operation
	mapCounter    int
	reduceCounter int
	done          chan bool
}

// Construct a new Master struct
func newMaster(address string) (master *Master) {
	master = new(Master)
	master.address = address
	master.done = make(chan bool)
	master.workers = make([]*RemoteWorker)
	master.idleWorkerChan = make(*RemoteWorker, IDLE_WORKER_BUFFER)
	master.totalWorkers = 0
	master.mapCounter = 0
	master.reduceCounter = 0
	return
}

// acceptMultipleConnections will handle the connections from multiple workers.
func (master *Master) acceptMultipleConnections() error {
	var (
		err     error
		newConn net.Conn
	)

	log.Printf("Accepting connections on %v\n", master.listener.Addr())

	for {
		newConn, err = master.listener.Accept()

		if err == nil {
			go master.handleConnection(&newConn)
		} else {
			log.Println("Failed to accept connection. Error: ", err)
			break
		}
	}

	log.Println("Stopped accepting connections.")
	return nil
}

// Handle a single connection until it's done, then closes it.
func (master *Master) handleConnection(conn *net.Conn) error {
	master.rpcServer.ServeConn(*conn)
	(*conn).Close()
	return nil
}
