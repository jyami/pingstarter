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

	// バックグラウンド実行終了後の終了メッセージを貯めるキュー
	// 10はとりあえずの値
	bgTerminateMsgs := make(chan string, 10)

	for {
		fmt.Printf("(pingstarter) > ")

		line, _, err := reader.ReadLine()
		if err != nil {
			log.Fatalf("ReadLine %v", err)
		}
		command := string(line)

		// "f"(フォアグラウンド)か"b"(バックグラウンド)を入力したらpingを実行する。
		//　それ以外の場合は何もしない
		if command == "f" || command == "b" {
			foreground := (command == "f")
			var procAttr os.ProcAttr
			procAttr.Sys = &syscall.SysProcAttr{Setpgid: true, Foreground: foreground}
			procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}

			process, err := os.StartProcess("/usr/bin/ping", []string{"/usr/bin/ping", "-c", "3", "yahoo.co.jp"}, &procAttr)
			if err != nil {
				log.Fatalf("StartProcess %v", err)
			}

			// 起動したpingの実行終了を待つ関数を定義
			waitCommandTerminated := func() {
				_, err = process.Wait()
				if err != nil {
					log.Fatalf("process.Wait %v", err)
				}

				if foreground {
				} else {
					bgTerminateMsgs <- "terminated."
				}
			}

			if foreground {
				// ブロック
				waitCommandTerminated()
			} else {
				// ノンブロック
				// 「起動したpingの実行終了を待つ」関数をゴルーチンで実行
				go waitCommandTerminated()
			}
		}

	L:
		for {
			select {
			// キューにメッセージが貯まっていれば、それを一つ取り出して
			case v := <-bgTerminateMsgs:
				fmt.Println(v)
			// キューにメッセージがなければ
			default:
				break L
			}
		}
	}
}
