package basher

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"bitbucket.org/kardianos/osext"
	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stderr)
	if os.Getenv("DEBUG") == "" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
}

func exitStatus(err error) (int, error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return int(status.ExitStatus()), nil
			}
		}
		return 0, err
	}
	return 0, nil
}

func RunBash(envfile string, command string, args []string, env []string) (int, error) {
	executable, err := osext.Executable()
	if err != nil {
		return 0, err
	}
	argstring := ""
	for _, arg := range args {
		argstring = argstring + " '" + arg + "'"
	}
	cmd := exec.Command("/usr/bin/env", "bash", "-c", command+argstring)
	cmd.Env = append(env,
		"PROGRAM="+executable,
		"BASH_ENV="+envfile,
	)

	if os.Getenv("DEBUG") != "" {
		fmt.Println("[go-basher] CMD", "/usr/bin/env -i ", strings.Join(cmd.Env, " "), "bash", "-c \"", command+argstring, "\"")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return exitStatus(cmd.Run())
}

type Context struct {
	sync.Mutex
	files []string
	env   []string
	fns   map[string]func([]string) int
}

func NewContext() *Context {
	return &Context{
		files: make([]string, 0),
		env:   make([]string, 0),
		fns:   make(map[string]func([]string) int),
	}
}

func NewContextWithEnv(e []string) *Context {
	return &Context{
		files: make([]string, 0),
		env:   e,
		fns:   make(map[string]func([]string) int),
	}
}

func (c *Context) Source(filename string) {
	c.Lock()
	defer c.Unlock()
	c.files = append(c.files, filename)
}

func (c *Context) Export(name string, value string) {
	c.Lock()
	defer c.Unlock()
	c.env = append(c.env, name+"="+value)
}

func (c *Context) ExportFunc(name string, fn func([]string) int) {
	c.Lock()
	defer c.Unlock()
	c.fns[name] = fn
}

func (c *Context) HandleFuncs(args []string) {
	log.Debug("HandleFuncs-hack() ")
	c.ExportFunc("help", func(a []string) int {
		prog := os.Args[0]
		fmt.Println("Built in functions:")
		for fn, _ := range c.fns {
			fmt.Println(prog, "::", fn)
		}
		return 0
	})

	for i, arg := range args {
		if arg == "::" && len(args) > i+1 {
			c.Lock()
			defer c.Unlock()
			for cmd := range c.fns {
				log.Debug("checking cmd: ", cmd)
				if cmd == args[i+1] {
					os.Exit(c.fns[cmd](args[i+2:]))
				}
			}
			os.Exit(6)
		}
	}
}

func (c *Context) buildEnvfile() (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "bashenv.")
	if err != nil {
		return "", err
	}
	defer file.Close()
	for _, filename := range c.files {
		f, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		defer f.Close()
		_, err = io.Copy(file, f)
		if err != nil {
			return "", err
		}
	}
	for cmd := range c.fns {
		file.Write([]byte(cmd + "() { $PROGRAM :: " + cmd + " \"$@\"; }\n"))
	}
	return file.Name(), nil
}

func (c *Context) Run(command string, args []string) (int, error) {
	c.Lock()
	defer c.Unlock()
	envfile, err := c.buildEnvfile()
	if err != nil {
		return 0, err
	}

	if os.Getenv("DEBUG") == "" {
		defer os.Remove(envfile)
	} else {
		log.Debug("[go-basher] keeping: ", envfile)
	}

	return RunBash(envfile, command, args, c.env)
}
