package internal

import (
	"os"
	"os/exec"
	"runtime"
)

// ClearScreen clears the terminal screen.
func ClearScreen() {
	if runtime.GOOS == "windows" {
		execCommand("cmd", "/c", "cls")
	} else {
		execCommand("clear")
	}
}

func execCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
