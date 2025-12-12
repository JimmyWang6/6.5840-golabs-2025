package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	mu      sync.Mutex
	Counter int
	Files   []File
	NReduce int
	Workers map[int]*WorkerItem
}

type File struct {
	Name     string
	Assigned bool
	Done     bool
	Owner    int
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) Register(args *RegisterArgs, reply *RegisterReply) error {
	_ = args
	c.mu.Lock()
	defer c.mu.Unlock()
	cur := c.Counter
	c.Counter++
	worker := &WorkerItem{WorkerID: cur, Files: make([]string, 0), LastSeen: time.Now()}
	// assign tasks while holding the lock
	assigned := assign(worker, c)
	if assigned.Name != "" {
		reply.File = assigned.Name
		// assign already appended filename into worker.Files inside assign
	}
	reply.WorkerID = cur
	c.Workers[cur] = worker
	return nil
}

// assign finds one unassigned file, marks it assigned to worker w,
// appends the filename to w.Files and returns that File value.
// returns zero File (Name == "") if no file available.
func assign(w *WorkerItem, c *Coordinator) File {
	var zero File
	if w == nil {
		return zero
	}
	for i := range c.Files {
		file := &c.Files[i]
		if !file.Assigned && !file.Done {
			file.Assigned = true
			file.Owner = w.WorkerID
			w.Files = append(w.Files, file.Name)
			// assign only one file at a time
			return *file
		}
	}
	return zero
}

func (c *Coordinator) allFinished() bool {
	for i := range c.Files {
		if c.Files[i].Done == false {
			return false
		}
	}
	return true
}

func (c *Coordinator) allProcessed() bool {
	for i := range c.Files {
		if c.Files[i].Assigned == false {
			return false
		}
	}
	return true
}

func (c *Coordinator) Report(args *ReportArgs, reply *ReportReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := args.WorkerID
	fileName := args.File

	// find and mark the file done if owned by this worker
	for i := range c.Files {
		file := &c.Files[i]
		if file.Name == fileName && file.Owner == id {
			file.Done = true
			file.Assigned = true
			log.Printf("Worker %d finished file %s\n", id, fileName)
			break
		}
	}
	if c.allFinished() {
		reply.Response = "done"
	}
	if c.allProcessed() {
		reply.Response = "processed"
	}
	// assign a new file if available
	// no need to check here because there must assign one file
	assigned := assign(c.Workers[id], c)
	reply.File = assigned.Name
	reply.Response = "doing"
	return nil
}

func (c *Coordinator) Heartbeat(args *HeartbeatArgs, reply *HeartbeatReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := args.WorkerID
	if w, ok := c.Workers[id]; ok {
		w.LastSeen = time.Now()
	}
	// TODO what happen if worker not found
	reply.Response = "ok"
	return nil
}

func (c *Coordinator) scanWorkers() {
	for {
		time.Sleep(time.Second * 2)
		c.mu.Lock()
		now := time.Now()
		for id, worker := range c.Workers {
			if now.Sub(worker.LastSeen).Seconds() > 10 {
				log.Printf("Worker %d is dead\n", worker.WorkerID)
				c.deleteWorker(id)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Coordinator) deleteWorker(id int) {
	worker := c.Workers[id]
	delete(c.Workers, id)
	for i := range c.Files {
		file := &c.Files[i]
		if file.Owner == worker.WorkerID && file.Done == false {
			file.Assigned = false
			file.Owner = -1
		}
	}
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	if err := rpc.Register(c); err != nil {
		log.Fatalf("rpc.Register error: %v", err)
	}
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	if err := os.Remove(sockname); err != nil && !os.IsNotExist(err) {
		log.Printf("remove socket error: %v", err)
	}
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		if err := http.Serve(l, nil); err != nil {
			log.Fatalf("http.Serve error: %v", err)
		}
	}()
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Files {
		if c.Files[i].Done == false {
			return false
		}
	}
	return true
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	newFile := make([]File, len(files))
	for i := range files {
		str := files[i]
		cur := File{str, false, false, -1}
		newFile[i] = cur
	}
	c := Coordinator{mu: sync.Mutex{}, Counter: 0, Files: newFile, NReduce: nReduce, Workers: make(map[int]*WorkerItem)}
	go c.scanWorkers()
	c.server()
	return &c
}
