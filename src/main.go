package main

import (
    "context"
    "fmt"
    "os/exec"
    "syscall"
    "time"
)

func main() {
    // Set the number of times to run the program and the timeout duration
    n := 10
    timeout := time.Second * 10

    // Run the program n times or until the timeout is reached
    for i := 0; i < n; i++ {
        if err := runProgramUntilTimeout(timeout); err != nil {
            fmt.Printf("Program timed out on iteration %d\n", i+1)
            break
        }
    }
}

func runProgramUntilTimeout(timeout time.Duration) error {
    // Create a context with the specified timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // Create a command to run the main function in main/main.go
    cmd := exec.CommandContext(ctx, "go", "run", "main/.")

    // Start the command
    if err := cmd.Start(); err != nil {
        return err
    }

    // Wait for the command to complete or for the context to be cancelled
    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
            return err
        }
        <-done
        return fmt.Errorf("program timed out")
    }
}