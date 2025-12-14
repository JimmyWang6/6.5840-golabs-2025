package mr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

type encFile struct {
	file *os.File
	enc  *json.Encoder
}

type WorkerItem struct {
	WorkerID int
	NReduce  int
	LastSeen time.Time
	File     File
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
	register(w)
	for {
		res := report(w)
		if res.Response == "done" {
			break
		}
		if res.Response == "wait" {
			time.Sleep(2 * time.Second)
			continue
		}
		w.File = res.File
		// here we would run the map or reduce task based on w.cur
		// for simplicity, we just print it out
		fmt.Printf("Worker %d processing file: %s\n", w.WorkerID, w.File.Name)
		if res.Type == StateMap {
			// run map task
			contents := Read(w.File.Name)
			prefix := fmt.Sprintf("mr-%d-", w.File.index)
			kva := mapf(w.File.Name, contents)

			// 为当前这个 map 任务按 reduce index 创建/覆盖中间文件
			encFiles := make([]encFile, w.NReduce)
			for i := 0; i < w.NReduce; i++ {
				fileName := fmt.Sprintf("%s%d", prefix, i)
				f, err := os.Create(fileName) // 始终重新创建（MapReduce 规范：每个任务自己的文件）
				if err != nil {
					log.Fatalf("cannot create %v", fileName)
				}
				defer f.Close()
				encFiles[i] = encFile{
					file: f,
					enc:  json.NewEncoder(f),
				}
			}
			for _, kv := range kva {
				r := ihash(kv.Key) % w.NReduce
				if err := encFiles[r].enc.Encode(&kv); err != nil {
					log.Fatalf("cannot encode kv: %v", err)
				}
			}
		} else if res.Type == StateReduce {
			fmt.Printf("Worker %d processing file: %s\n", w.WorkerID, w.File.Name)
		}
	}
}

func Read(name string) string {
	f, err := os.Open(name)
	if err != nil {
		log.Fatalf("cannot open %v", name)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("cannot read %v", name)
	}
	return string(data)
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func register(w *WorkerItem) error {
	args := RegisterArgs{}
	reply := RegisterReply{}

	ok := call("Coordinator.Register", &args, &reply)
	if ok {
		// use reply.WorkerID and reply.UniqueKey
		w.WorkerID = reply.WorkerID
		w.NReduce = reply.NReduce
	} else {
		log.Fatalf("Register call failed!\n")
	}
	return nil
}

func report(w *WorkerItem) ReportReply {
	args := ReportArgs{}
	if &w.File == nil {
		// first report, no file assigned yet
		log.Printf("worker %d has no file", w.WorkerID)
		args = ReportArgs{WorkerID: w.WorkerID, File: -1}
	} else {
		args = ReportArgs{WorkerID: w.WorkerID, File: w.File.index}
	}
	reply := ReportReply{}
	call("Coordinator.Report", &args, &reply)
	return reply
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
