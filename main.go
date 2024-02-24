package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"
)

func main() {

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("(pingstarter) > ")

		_, _, err := reader.ReadLine()
		if err != nil {
			log.Fatalf("ReadLine %v", err)
		}

		var procAttr os.ProcAttr
		procAttr.Sys = &syscall.SysProcAttr{Setpgid: true}
		procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

		process, err := os.StartProcess("/usr/bin/ping", []string{"/usr/bin/ping", "-c", "3", "yahoo.co.jp"}, &procAttr)
		if err != nil {
			log.Fatalf("StartProcess %v", err)
		}

		_, err = process.Wait()
		if err != nil {
			log.Fatalf("Wait %v", err)
		}
	}
}
