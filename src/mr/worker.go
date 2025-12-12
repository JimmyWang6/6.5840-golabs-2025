package mr

import (
	"fmt"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type WorkerItem struct {
	WorkerID int
	Files    []string
	LastSeen time.Time
	cur      string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// Example RPC demo (keeps existing example behavior).
	// In the real lab you should implement a loop that contacts the
	// coordinator, fetches Map/Reduce tasks, runs them and reports back.
	// See runMapTask below for how to execute mapf on a given file.
	var w = &WorkerItem{}
	Register(w)
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func Register(w *WorkerItem) error {
	args := RegisterArgs{}
	reply := RegisterReply{}

	ok := call("Coordinator.Register", &args, &reply)
	if ok {
		// use reply.WorkerID and reply.UniqueKey
		w.WorkerID = reply.WorkerID
		w.cur = reply.File
	} else {
		log.Fatalf("Register call failed!\n")
	}
	return nil
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
