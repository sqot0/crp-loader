package terminal

import (
	"fmt"
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

func DisplayLogo() {
	fmt.Println("   ____ ____  ____    _     ___    _    ____  _____ ____")
	fmt.Println("  / ___|  _ \\|  _ \\  | |   / _ \\  / \\  |  _ \\| ____|  _ \\")
	fmt.Println(" | |   | |_) | |_) | | |  | | | |/ _ \\ | | | |  _| | |_) |")
	fmt.Println(" | |___|  _ <|  __/  | |__| |_| / ___ \\| |_| | |___|  _ <")
	fmt.Println("  \\____|_| \\_\\_|     |_____\\___/_/   \\_\\____/|_____|_| \\_\\")
	fmt.Print("\n\n")
}

func execCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
