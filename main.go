package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("(pingstarter) > ")

		line, _, err := reader.ReadLine()
		if err != nil {
			log.Fatalf("ReadLine %v", err)
		}

		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

		process, err := os.StartProcess(string(line), []string{string(line)}, &procAttr)
		if err != nil {
			log.Fatalf("StartProcess %v", err)
		}

		_, err = process.Wait()
		if err != nil {
			log.Fatalf("Wait %v", err)
		}
	}
}
