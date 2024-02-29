package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {

	signal.Ignore(syscall.SIGTTIN, syscall.SIGTTOU)

	pgrpid := tcgetpgrp()

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
					tcsetpgrp(pgrpid)
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

// 現在のプロセスの制御端末を取得
func devTty() *os.File {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		log.Fatalf("Couldn't open /dev/tty %s", err)
	}
	return tty
}

// 指定した制御端末のフォアグラウンドプロセスグループIDを取得
func tcgetpgrp() int {
	tty := devTty()
	defer tty.Close()

	pgrpid, _ := unix.IoctlGetInt(int(tty.Fd()), unix.TIOCGPGRP)
	return pgrpid
}

// 指定したプロセスグループを指定した制御端末のフォアグラウンドプロセスグループにする。
func tcsetpgrp(pgrpid int) {
	tty := devTty()
	defer tty.Close()

	unix.IoctlSetPointerInt(int(tty.Fd()), unix.TIOCSPGRP, pgrpid)
}
