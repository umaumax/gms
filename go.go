package main

import (
	"os/exec"
)

func goexec(name string) (b []byte, err error) {
	cmd := exec.Command("go", "run", name)
	return cmd.CombinedOutput()
}
