package samples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/samples/taskboard/task"
)

func TestSamplesCompile(t *testing.T) {
	command := exec.Command("go", "test", "./samples/api", "./samples/taskboard")
	command.Dir = repoRoot(t)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("samples should compile: %v\n%s", err, output)
	}
}

func TestSamplesCompileFromSampleDirectories(t *testing.T) {
	root := repoRoot(t)
	for _, sample := range []string{"api", "taskboard"} {
		t.Run(sample, func(t *testing.T) {
			command := exec.Command("go", "test", ".")
			command.Dir = filepath.Join(root, "samples", sample)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("sample should compile from its directory: %v\n%s", err, output)
			}
		})
	}
}

func TestTaskboardDefaultSQLitePathWorksFromSampleDirectory(t *testing.T) {
	t.Setenv("SQLITE_PATH", "")

	sampleDir := filepath.Join(repoRoot(t), "samples", "taskboard")
	databasePath := filepath.Join(sampleDir, "nexgou.db")
	_ = os.Remove(databasePath)
	t.Cleanup(func() { _ = os.Remove(databasePath) })

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })
	if err := os.Chdir(sampleDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	config := nexgou.ConfigService{}
	log := nexgou.NewLogger(nexgou.LoggerOptions{Level: nexgou.LevelSilent})
	store := task.NewStore(&config, log)
	defer store.Close()

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("default sqlite database was not created: %v", err)
	}
}

func TestTaskboardCreatesSQLiteParentDirectory(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "taskboard.db")
	t.Setenv("SQLITE_PATH", databasePath)

	config := nexgou.ConfigService{}
	log := nexgou.NewLogger(nexgou.LoggerOptions{Level: nexgou.LevelSilent})
	store := task.NewStore(&config, log)
	defer store.Close()

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("sqlite database was not created: %v", err)
	}
}

func TestTaskboardPersistsTasksInSQLite(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "taskboard.db")
	t.Setenv("SQLITE_PATH", databasePath)

	config := nexgou.ConfigService{}
	log := nexgou.NewLogger(nexgou.LoggerOptions{Level: nexgou.LevelSilent})
	store := task.NewStore(&config, log)
	defer store.Close()

	service := task.NewService(store, log)
	created, err := service.Create("Write migration tests", "alice")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	tasks, err := service.FindAll("alice")
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title != created.Title || tasks[0].UserID != "alice" {
		t.Fatalf("tasks = %+v, want persisted task", tasks)
	}

	completed, err := service.Complete(toID(created.ID))
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if !completed.Done {
		t.Fatalf("completed task = %+v, want done", completed)
	}

	if err := service.Delete(toID(created.ID)); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	tasks, err = service.FindAll("alice")
	if err != nil {
		t.Fatalf("FindAll after delete returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("tasks = %+v, want empty list after delete", tasks)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot find caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func toID(id int64) string {
	return strconv.FormatInt(id, 10)
}
