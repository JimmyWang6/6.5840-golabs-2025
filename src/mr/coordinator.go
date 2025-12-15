package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	StateReduce = "reduce"
	StateMap    = "map"
	StateFinish = "finish"
)

type Coordinator struct {
	mu      sync.Mutex
	State   string
	Counter int
	Files   []File
	NReduce int
	Workers map[int]*WorkerItem
}

type File struct {
	Name     string
	Index    int
	Assigned bool
	Done     bool
	Owner    int
	Type     string // map or reduce
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
	cur := c.Counter
	c.Counter++
	worker := &WorkerItem{WorkerID: cur, LastSeen: time.Now()}
	// assign tasks while holding the lock
	reply.WorkerID = cur
	reply.NReduce = c.NReduce
	count := 0
	for i := range c.Files {
		if c.Files[i].Type == StateMap {
			count++
		}
	}
	reply.NMap = count
	c.Workers[cur] = worker
	return nil
}

// assign finds one unassigned file, marks it assigned to worker w,
// appends the filename to w.Files and returns that File value.
// returns zero File (Name == "") if no file available.
func assign(w *WorkerItem, c *Coordinator) File {
	if c.State == StateMap {
		allDone := true
		// find first unassigned map file
		for i := range c.Files {
			file := &c.Files[i]
			if file.Type == StateMap {
				if !file.Done && !file.Assigned {
					file.Assigned = true
					file.Owner = w.WorkerID
					return *file
				}
				if !file.Done {
					allDone = false
				}
			}
		}
		if allDone {
			log.Printf("Map phase finished, start reduce phase\n")
			c.State = StateReduce
			return doReduce(w, c)
		}
		return File{}
	}
	return doReduce(w, c)
}

func doReduce(w *WorkerItem, c *Coordinator) File {
	for i := range c.Files {
		file := &c.Files[i]
		if file.Type == StateReduce {
			if !file.Done && !file.Assigned {
				file.Assigned = true
				file.Owner = w.WorkerID
				return *file
			}
		}
	}
	allDone := true
	for i := range c.Files {
		if c.Files[i].Type == StateReduce && !c.Files[i].Done {
			allDone = false
			break
		}
	}
	if allDone {
		log.Printf("All tasks finished\n")
		c.State = StateFinish
	}
	return File{}
}

func (c *Coordinator) allFinished() bool {
	return c.State == StateFinish
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
	log.Printf("Report received from worker %d for file index %d\n", args.WorkerID, args.File)
	defer log.Printf("Finished report for file index %d\n", args.File)
	id := args.WorkerID
	fileIndex := args.File
	if fileIndex == -1 {
		// empty file name means first time request
		log.Printf("File not found in worker %d\n", args.WorkerID)
		assigned := assign(c.Workers[id], c)
		if assigned.Name == "" {
			if c.allFinished() {
				reply.Response = "done"
			} else {
				reply.Response = "wait"
				log.Printf("replying wait to worker, reply %d\n", args.WorkerID)
			}
			reply.File = File{}
			return nil
		} else {
			reply.File = assigned
			reply.Response = "doing"
			return nil
		}
	}

	// find and mark the file done if owned by this worker
	for i := range c.Files {
		file := &c.Files[i]
		if file.Index == fileIndex && file.Owner == id && !file.Done {
			log.Printf("File %s is already processed\n", file.Name)
			file.Done = true
			file.Assigned = true
			break
		}
	}
	// assign a new file if available
	// no need to check here because there must assign one file
	assigned := assign(c.Workers[id], c)
	// TODO duplicate remove
	if assigned.Name == "" {
		if c.allFinished() {
			reply.Response = "done"
		} else {
			reply.Response = "wait"
			reply.File = File{Index: -1}
		}
		return nil
	}
	reply.File = assigned
	reply.Response = "doing"
	return nil
}

func (c *Coordinator) Heartbeat(args *HeartbeatArgs, reply *HeartbeatReply) error {
	id := args.WorkerID
	if w, ok := c.Workers[id]; ok {
		w.LastSeen = time.Now()
	}
	log.Printf("Heartbeat received from worker %d\n", id)
	// TODO what happen if worker not found
	reply.Response = "ok"
	return nil
}

func (c *Coordinator) scanWorkers() {
	for {
		time.Sleep(time.Second * 10)
		now := time.Now()
		for id, worker := range c.Workers {
			if now.Sub(worker.LastSeen).Seconds() > 10 {
				log.Printf("Worker %d is dead\n", worker.WorkerID)
				c.deleteWorker(id)
			}
		}
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
	// 统计未完成任务
	mapTotal, mapDone, mapAssigned := 0, 0, 0
	reduceTotal, reduceDone, reduceAssigned := 0, 0, 0

	for i := range c.Files {
		if c.Files[i].Type == StateMap {
			mapTotal++
			if c.Files[i].Done {
				mapDone++
			}
			if c.Files[i].Assigned {
				mapAssigned++
			}
		} else if c.Files[i].Type == StateReduce {
			reduceTotal++
			if c.Files[i].Done {
				reduceDone++
			}
			if c.Files[i].Assigned {
				reduceAssigned++
			}
		}
	}
	isDone := c.allFinished()
	log.Printf("========== Done() Check ==========\n")
	log.Printf("State: %s\n", c.State)
	log.Printf("Map    : Total=%d, Done=%d, Assigned=%d\n", mapTotal, mapDone, mapAssigned)
	log.Printf("Reduce : Total=%d, Done=%d, Assigned=%d\n", reduceTotal, reduceDone, reduceAssigned)
	log.Printf("AllFinished: %v\n", isDone)
	log.Printf("===================================\n")
	return c.allFinished()
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	nMap := len(files)
	total := nMap + nReduce
	newFile := make([]File, total)

	for i := 0; i < nMap; i++ {
		newFile[i] = File{
			Name:     files[i],
			Index:    i,
			Assigned: false,
			Done:     false,
			Owner:    -1,
			Type:     StateMap,
		}
	}

	for k := 0; k < nReduce; k++ {
		idx := k + nMap
		newFile[idx] = File{
			Name:     strconv.Itoa(k), // reduce task id as name
			Index:    idx,
			Assigned: false,
			Done:     false,
			Owner:    -1,
			Type:     StateReduce,
		}
	}

	c := Coordinator{
		mu:      sync.Mutex{},
		State:   StateMap,
		Counter: 0,
		Files:   newFile,
		NReduce: nReduce,
		Workers: make(map[int]*WorkerItem),
	}
	go c.scanWorkers()
	c.server()
	return &c
}
