package main

import (
	"bufio"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/pkg/term/termios"
)

// Commander manages multiple commands based on a template and args.
type Commander struct {
	Commands []*Command
	template []string
	tags     []string
	wg       sync.WaitGroup
}

// Command wraps an exec.Cmd with its output and tag.
type Command struct {
	ExitCode int
	Cmd      *exec.Cmd
	Bytes    []byte
	Tag      string
}

// NewCommanderFromArgs parses the args and returns a commander object ready to
// run one command for each tag.
func NewCommanderFromArgs(args []string) *Commander {
	c := &Commander{
		template: make([]string, 0),
		tags:     make([]string, 0),
	}

	i := 0
	for ; i < len(args); i++ {
		if args[i] == ":::" {
			i++
			break
		}
		c.template = append(c.template, args[i])
	}
	// TODO: or get from stdin
	for ; i < len(args); i++ {
		c.tags = append(c.tags, args[i])
	}

	c.Commands = make([]*Command, len(c.tags))
	return c
}

// Signal sends the provided signal to each command.
func (c *Commander) Signal(sig os.Signal) {
	for _, cmd := range c.Commands {
		if cmd.ExitCode == -1 {
			syscall.Kill(-cmd.Cmd.Process.Pid, sig.(syscall.Signal))
		}
	}
}

// Start starts all the commands in the background.
func (c *Commander) Start() {
	for i, tag := range c.tags {
		args := make([]string, len(c.template))
		copy(args, c.template)
		found := false
		for i, a := range args {
			if a == "{}" {
				found = true
				args[i] = tag
			}
		}
		if !found {
			args = append(args, tag)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		// TODO: arg for no-tty
		m, s, err := termios.Pty()
		if err != nil {
			// FIXME
			panic(err)
		}
		// cmd.Stdin = m
		cmd.Stdout = s
		cmd.Stderr = s
		err = cmd.Start()
		if err != nil {
			// FIXME
			panic(err)
		}

		c.Commands[i] = &Command{
			ExitCode: -1,
			Cmd:      cmd,
			Bytes:    make([]byte, 0),
			Tag:      tag,
		}

		c.wg.Add(1)
		go func(cc *Command) {
			defer func() {
				m.Close()
				// This close should end the scanner.
				s.Close()
			}()
			cmd.Wait()
			cc.ExitCode = cmd.ProcessState.ExitCode()
		}(c.Commands[i])

		go func(cc *Command) {
			defer c.wg.Done()
			scanner := bufio.NewScanner(m)
			scanner.Split(bufio.ScanBytes)
			for scanner.Scan() {
				b := scanner.Bytes()
				cc.Bytes = append(cc.Bytes, b...)
			}
		}(c.Commands[i])
	}
}

// Wait waits for all commands to finish.
func (c *Commander) Wait() {
	c.wg.Wait()
}

// IsDone returns true if all commands have finished.
func (c *Commander) IsDone() bool {
	for _, cmd := range c.Commands {
		if cmd.ExitCode == -1 {
			return false
		}
	}
	return true
}
