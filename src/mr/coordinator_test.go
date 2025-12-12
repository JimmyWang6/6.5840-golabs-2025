package mr

import (
	"testing"
)

func TestCoordinator(t *testing.T) {
	coordinator := MakeCoordinator([]string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt", "file6.txt"}, 5)
	if coordinator == nil {
		t.Errorf("Failed to create coordinator")
	}
	ra1 := RegisterArgs{}
	rs1 := RegisterReply{}
	coordinator.Register(&ra1, &rs1)
	// Get file 1 assigned to worker 0
	if rs1.WorkerID != 0 {
		t.Errorf("Expected WorkerID 0, got %d", rs1.WorkerID)
	}
	if rs1.File != "file1.txt" {
		t.Errorf("Expected assigned file 'file1.txt', got %s", rs1.File)
	}
	ra2 := RegisterArgs{}
	rs2 := RegisterReply{}
	coordinator.Register(&ra2, &rs2)
	if rs2.WorkerID != 1 {
		t.Errorf("Expected WorkerID 1, got %d", rs2.WorkerID)
	}
	if rs2.File != "file2.txt" {
		t.Errorf("Expected assigned file 'file2.txt', got %s", rs2.File)
	}
	if len(coordinator.Workers) != 2 {
		t.Errorf("Expected 2 workers registered, got %d", len(coordinator.Workers))
	}
	rpArgs1 := ReportArgs{0, rs1.File}
	rpReply1 := ReportReply{}
	coordinator.Report(&rpArgs1, &rpReply1)
	if rpReply1.File != "file3.txt" {
		t.Errorf("Expected next assigned file 'file3.txt', got %s", rpReply1.File)
	}

	// delete worker1 (due to timeout or something)
	coordinator.deleteWorker(0)

	// worker 1 report back (late)
	rpArgs2 := ReportArgs{1, rs2.File}
	rpReply2 := ReportReply{}
	coordinator.Report(&rpArgs2, &rpReply2)

	// should get file3
	if rpReply2.File != "file3.txt" {
		t.Errorf("Expected reassigned file 'file3.txt', got %s", rpReply2.File)
	}
}
