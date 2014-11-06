package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("/usr/bin/env", "bash", "-c", "main")
	var env []string
	cmd.Env = append(env,
		"BASH_ENV=./test.bash",
	)

	fmt.Printf("exec.Cmd: %#v\n", cmd)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output: ", string(out))
}
