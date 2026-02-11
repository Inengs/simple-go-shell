package jobs // this handles background process management
import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Job struct {
	ID int
	PID int
	CommandString string
	Status string 
}

var jobsMap = []Job{}
var nextJobID = 1

func ExecuteCommandWithJobs(args []string) error{
	if len(args) == 0 {
		return fmt.Errorf("input is empty")
	}

	lastArgument := args[len(args)-1]

	isBackground := false

	if lastArgument == "&" { // check if the last argument is &
		isBackground = true // set a flag that it is a background job
		args = args[:len(args)-1] // ONLY remove & if it exists
	}

	// Check again after removing &
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	command := args[0] // Extract the command name from the first element

	path, err := exec.LookPath(command) // Searches your system's PATH for the executable

	if err != nil { // If the command wasn't found, return the error to the caller
		return fmt.Errorf("command not found: %s", command)
	}

	cmd := exec.Command(path, args[1:]...)

	// Connect stdin, stdout, stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if isBackground {
		err := cmd.Start() // start and check error

		if err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}
		PID := cmd.Process.Pid

		// Create job with nextJobID, not hardcoded 1
		job := Job{
			ID:            nextJobID,
			PID:           PID,
			CommandString: strings.Join(args, " "), // Join args as string, not cmd object
			Status:        "running",
		}

		jobsMap = append(jobsMap, job) // Add job to the map
		
		// Print job notification
		fmt.Printf("[%d] %d\n", job.ID, job.PID)
		
		nextJobID++ // Increment for next job
	} else {
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	return nil
}

// Builtin function to list all jobs
func ListJobs() {
	if len(jobsMap) == 0 {
		fmt.Println("No background jobs")
		return
	}

	for _, job := range jobsMap {
		fmt.Printf("[%d]  %s    %s\n", job.ID, job.Status, job.CommandString)
	}
}

// Add these to your jobs package

func UpdateJobStatus(jobID int, status string) {
    for i := range jobsMap {
        if jobsMap[i].ID == jobID {
            jobsMap[i].Status = status
            return
        }
    }
}

func RemoveCompletedJobs() {
    var activeJobs []Job
    
    for _, job := range jobsMap {
        // Check if process still exists
        process, err := os.FindProcess(job.PID)
        if err != nil {
            // Process not found, skip it
            continue
        }
        
        // Try to signal the process (signal 0 doesn't actually send a signal, just checks if alive)
        err = process.Signal(syscall.Signal(0))
        if err != nil {
            // Process finished, try to wait on it to clean up zombie
            process.Wait()
            fmt.Printf("[%d]  Done    %s\n", job.ID, job.CommandString)
            continue
        }
        
        // Process still running, keep it
        activeJobs = append(activeJobs, job)
    }
    
    jobsMap = activeJobs
}

func GetJobByID(id int) (*Job, error) {
    for i := range jobsMap {
        if jobsMap[i].ID == id {
            return &jobsMap[i], nil
        }
    }
    return nil, fmt.Errorf("job [%d] not found", id)
}