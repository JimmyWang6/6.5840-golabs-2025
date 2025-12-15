package mr

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
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
	NMap     int
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
	go heartbeat(w)
	for {
		res := report(w)
		log.Printf("Worker %d: received response: %+v\n", w.WorkerID, res)
		if res.Response == "done" {
			break
		}
		if res.Response == "wait" {
			// no task assigned, wait and retry
			w.File = File{Index: -1}
			time.Sleep(2 * time.Second)
			continue
		}
		w.File = res.File
		// here we would run the map or reduce task based on w.cur
		// for simplicity, we just print it out
		if w.File.Type == StateMap {
			fmt.Printf("Worker %d: starting map task %d on file %s\n", w.WorkerID, w.File.Index, w.File.Name)
			// run map task
			contents := Read(w.File.Name)
			prefix := fmt.Sprintf("mr-%d-", w.File.Index)
			kva := mapf(w.File.Name, contents)

			// 为当前这个 map 任务按 reduce index 创建/覆盖中间文件
			encFiles := make([]encFile, w.NReduce)
			log.Printf("Worker %d: map task %d creating %d intermediate files\n", w.WorkerID, w.File.Index, w.NReduce)
			for i := 0; i < w.NReduce; i++ {
				fileName := fmt.Sprintf("%s%d", prefix, i)
				f, err := os.Create(fileName) // 始终重新创建（MapReduce 规范：每个任务自己的文件）
				if err != nil {
					log.Fatalf("cannot create %v", fileName)
				}
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
			for _, ef := range encFiles {
				if err := ef.file.Close(); err != nil {
					log.Fatalf("cannot close file: %v", err)
				}
			}
			log.Printf("Worker %d: map task %d produced %d key/value pairs\n", w.WorkerID, w.File.Index, len(kva))
		} else if w.File.Type == StateReduce {
			reduceIndex := w.File.Index - w.NMap
			intermediate := []KeyValue{}
			fmt.Printf("Worker %d: starting reduce task %d\n", w.WorkerID, reduceIndex)
			for i := 0; i < w.NMap; i++ {
				intermediateFileName := fmt.Sprintf("mr-%d-%d", i, reduceIndex)
				_, err := os.Stat(intermediateFileName)
				if os.IsNotExist(err) {
					log.Fatalf("intermediate file %s does not exist", intermediateFileName)
				}
				f, _ := os.Open(intermediateFileName)
				defer f.Close()
				dec := json.NewDecoder(f)
				for {
					var kv KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					intermediate = append(intermediate, kv)
				}
			}
			sort.Slice(intermediate, func(i, j int) bool {
				return intermediate[i].Key < intermediate[j].Key
			})
			out := fmt.Sprintf("mr-out-%d", reduceIndex)
			file, _ := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
			defer file.Close()
			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)
				result := fmt.Sprintf("%v %v\n", intermediate[i].Key, output)
				file.WriteString(result)
				i = j
			}
		}
		log.Printf("Worker %d: completed task %+v\n", w.WorkerID, w.File)
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

func heartbeat(w *WorkerItem) {
	args := HeartbeatArgs{WorkerID: w.WorkerID}
	reply := HeartbeatReply{}

	for {
		time.Sleep(3 * time.Second)
		ok := call("Coordinator.Heartbeat", &args, &reply)
		if ok {
			w.LastSeen = time.Now()
		} else {
			log.Fatalf("Heartbeat call failed!\n")
		}
	}
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
		w.NMap = reply.NMap
	} else {
		log.Fatalf("Register call failed!\n")
	}
	return nil
}

func report(w *WorkerItem) ReportReply {
	args := ReportArgs{}
	if w.File.Name == "" {
		// first report, no file assigned yet
		args = ReportArgs{WorkerID: w.WorkerID, File: -1}
	} else {
		args = ReportArgs{WorkerID: w.WorkerID, File: w.File.Index}
	}
	reply := ReportReply{}
	log.Printf("Worker %d: reporting completion of file %d\n", w.WorkerID, args.File)
	call("Coordinator.Report", &args, &reply)
	log.Printf("Report success!")
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
