package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// CommandRunner interface that wraps the RunWithoutRetry function.
type CommandRunner interface {
	RunWithoutRetry(c *Command) (string, error)
}

// Command is a struct containing the details of an external command to be executed.
type Command struct {
	attempts int
	Errors   []error
	Dir      string
	Name     string
	Args     []string
	Out      io.Writer
	Err      io.Writer
	In       io.Reader
	Env      map[string]string
}

// CommandError is the error object encapsulating an error from a Command.
type CommandError struct {
	Command Command
	Output  string
	cause   error
}

func (c CommandError) Error() string {
	// sanitise any password arguments before printing the error string. The actual sensitive argument is still present
	// in the Command object
	sanitisedArgs := make([]string, len(c.Command.Args))
	copy(sanitisedArgs, c.Command.Args)
	for i, arg := range sanitisedArgs {
		if strings.Contains(strings.ToLower(arg), "password") && i < len(sanitisedArgs)-1 {
			// sanitise the subsequent argument to any 'password' fields
			sanitisedArgs[i+1] = "*****"
		}
	}

	return fmt.Sprintf("failed to run '%s %s' command in directory '%s', output: '%s'",
		c.Command.Name, strings.Join(sanitisedArgs, " "), c.Command.Dir, c.Output)
}

// SetName Setter method for Name to enable use of interface instead of Command struct.
func (c *Command) SetName(name string) {
	c.Name = name
}

// CurrentName returns the current name of the command.
func (c *Command) CurrentName() string {
	return c.Name
}

// SetDir Setter method for Dir to enable use of interface instead of Command struct.
func (c *Command) SetDir(dir string) {
	c.Dir = dir
}

// CurrentDir returns the current Dir.
func (c *Command) CurrentDir() string {
	return c.Dir
}

// SetArgs Setter method for Args to enable use of interface instead of Command struct.
func (c *Command) SetArgs(args []string) {
	c.Args = args
}

// CurrentArgs returns the current command arguments.
func (c *Command) CurrentArgs() []string {
	return c.Args
}

// SetEnv Setter method for Env to enable use of interface instead of Command struct.
func (c *Command) SetEnv(env map[string]string) {
	c.Env = env
}

// CurrentEnv returns the current environment variables.
func (c *Command) CurrentEnv() map[string]string {
	return c.Env
}

// SetEnvVariable sets an environment variable into the environment.
func (c *Command) SetEnvVariable(name string, value string) {
	if c.Env == nil {
		c.Env = map[string]string{}
	}
	c.Env[name] = value
}

// Attempts The number of times the command has been executed.
func (c *Command) Attempts() int {
	return c.attempts
}

// DidError returns a boolean if any error occurred in any execution of the command.
func (c *Command) DidError() bool {
	return len(c.Errors) > 0
}

// DidFail returns a boolean if the command could not complete (errored on every attempt).
func (c *Command) DidFail() bool {
	return len(c.Errors) == c.attempts
}

// Error returns the last error.
func (c *Command) Error() error {
	if len(c.Errors) > 0 {
		return c.Errors[len(c.Errors)-1]
	}
	return nil
}

func (c *Command) String() string {
	var builder strings.Builder
	for k, v := range c.Env {
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(v)
		builder.WriteString(" ")
	}
	builder.WriteString(c.Name)
	for _, arg := range c.Args {
		builder.WriteString(" ")
		builder.WriteString(arg)
	}
	return builder.String()
}

type DefaultCommandRunner struct {
}

// RunWithoutRetry Execute the command without retrying on failure and block waiting for return values.
func (d DefaultCommandRunner) RunWithoutRetry(c *Command) (string, error) {
	var r string
	var e error

	r, e = d.run(c)
	c.attempts++
	if e != nil {
		c.Errors = append(c.Errors, e)
	}
	return r, e
}

func (d *DefaultCommandRunner) run(c *Command) (string, error) {
	e := exec.Command(c.Name, c.Args...) // #nosec
	if c.Dir != "" {
		e.Dir = c.Dir
	}
	if len(c.Env) > 0 {
		m := map[string]string{}
		environ := os.Environ()
		for _, kv := range environ {
			paths := strings.SplitN(kv, "=", 2)
			if len(paths) == 2 {
				m[paths[0]] = paths[1]
			}
		}
		for k, v := range c.Env {
			m[k] = v
		}
		envVars := []string{}
		for k, v := range m {
			envVars = append(envVars, k+"="+v)
		}
		e.Env = envVars
	}

	if c.Out != nil {
		e.Stdout = c.Out
	}

	if c.Err != nil {
		e.Stderr = c.Err
	}

	if c.In != nil {
		e.Stdin = c.In
	}

	var text string
	var err error

	if c.Out != nil {
		err := e.Run()
		if err != nil {
			return text, CommandError{
				Command: *c,
				cause:   err,
			}
		}
	} else {
		data, err := e.CombinedOutput()
		output := string(data)
		text = strings.TrimSpace(output)
		if err != nil {
			return text, CommandError{
				Command: *c,
				Output:  text,
				cause:   err,
			}
		}
	}

	return text, err
}
