package mr

import (
	"testing"
)

func TestCoordinator(t *testing.T) {
	coordinator := MakeCoordinator([]string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt", "file6.txt"}, 5)
	if coordinator == nil {
		t.Errorf("Failed to create coordinator")
	}

	// register first worker
	ra1 := RegisterArgs{}
	rs1 := RegisterReply{}
	coordinator.Register(&ra1, &rs1)
	// worker 0 should get WorkerID 0
	if rs1.WorkerID != 0 {
		t.Errorf("Expected WorkerID 0, got %d", rs1.WorkerID)
	}

	// register second worker
	ra2 := RegisterArgs{}
	rs2 := RegisterReply{}
	coordinator.Register(&ra2, &rs2)
	if rs2.WorkerID != 1 {
		t.Errorf("Expected WorkerID 1, got %d", rs2.WorkerID)
	}

	// coordinator should track two workers
	if len(coordinator.Workers) != 2 {
		t.Errorf("Expected 2 workers registered, got %d", len(coordinator.Workers))
	}

	// worker 0 asks for a task (no previous file to report)
	rpArgs1 := ReportArgs{WorkerID: 0, File: ""}
	rpReply1 := ReportReply{}
	coordinator.Report(&rpArgs1, &rpReply1)
	// first assigned file should be file1.txt
	if rpReply1.File != "file1.txt" {
		t.Errorf("Expected first assigned file 'file1.txt', got %s", rpReply1.File)
	}

	// worker 0 reports completion of file1 and asks for next task
	rpArgs2 := ReportArgs{WorkerID: 0, File: rpReply1.File}
	rpReply2 := ReportReply{}
	coordinator.Report(&rpArgs2, &rpReply2)
	// next file in sequence should be file3.txt (since file2 goes to worker1 later)
	if rpReply2.File != "file3.txt" {
		t.Errorf("Expected next assigned file 'file3.txt', got %s", rpReply2.File)
	}

	// delete worker 0 (due to timeout or something)
	coordinator.deleteWorker(0)

	// worker 1 now asks for/reports its previous task; simulate it having processed file2
	// first, worker 1 requests its first file
	w1FirstArgs := ReportArgs{WorkerID: 1, File: ""}
	w1FirstReply := ReportReply{}
	coordinator.Report(&w1FirstArgs, &w1FirstReply)

	// worker 1 then reports completion of that file and asks for another task
	lateArgs := ReportArgs{WorkerID: 1, File: w1FirstReply.File}
	lateReply := ReportReply{}
	coordinator.Report(&lateArgs, &lateReply)

	// after worker 0 is deleted, the next unprocessed file that should be (re)assigned is file3.txt
	if lateReply.File != "file3.txt" {
		t.Errorf("Expected reassigned file 'file3.txt', got %s", lateReply.File)
	}
}
