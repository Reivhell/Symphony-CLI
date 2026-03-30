package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/Reivhell/symphony/internal/blueprint"
)

// RunHook runs a shell hook with cancellation support.
func RunHook(execCtx context.Context, action blueprint.Action, ctx *EngineContext) error {
	if action.Type != "shell" {
		return nil
	}

	commandStr, err := RenderString(action.Command, ctx)
	if err != nil {
		return fmt.Errorf("render hook command: %w", err)
	}

	wdStr := action.WorkingDir
	if wdStr != "" {
		wdStr, err = RenderString(wdStr, ctx)
		if err != nil {
			return fmt.Errorf("render hook working_dir: %w", err)
		}
	} else {
		wdStr = ctx.OutputDir
	}

	if ctx.Reporter != nil {
		ctx.Reporter.OnHookStart(commandStr)
	}

	var cmdExe *exec.Cmd
	if runtime.GOOS == "windows" {
		shell := os.Getenv("ComSpec")
		if shell == "" {
			shell = `C:\Windows\System32\cmd.exe`
		}
		cmdExe = exec.CommandContext(execCtx, shell, "/C", commandStr)
	} else {
		cmdExe = exec.CommandContext(execCtx, "sh", "-c", commandStr)
	}

	cmdExe.Dir = wdStr

	outPipe, _ := cmdExe.StdoutPipe()
	errPipe, _ := cmdExe.StderrPipe()

	if err := cmdExe.Start(); err != nil {
		return fmt.Errorf("hook start failed: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			if ctx.Reporter != nil {
				ctx.Reporter.OnHookOutput(scanner.Text())
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(errPipe)
		for scanner.Scan() {
			if ctx.Reporter != nil {
				ctx.Reporter.OnHookOutput("[stderr] " + scanner.Text())
			}
		}
	}()

	if err := cmdExe.Wait(); err != nil {
		if execCtx.Err() != nil {
			_, _ = fmt.Fprintf(os.Stderr, "symphony: hook interrupted; you may need to run manual cleanup (e.g. go mod tidy) in the output directory.\n")
		}
		return fmt.Errorf("hook failed: %w", err)
	}

	return nil
}
