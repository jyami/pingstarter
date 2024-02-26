package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var signals = []os.Signal{
	syscall.SIGABRT,
	syscall.SIGALRM,
	syscall.SIGBUS,
	//syscall.SIGCHLD, 通常は無視
	//syscall.SIGCLD, 通常は無視
	syscall.SIGCONT,
	syscall.SIGFPE,
	syscall.SIGHUP,
	syscall.SIGILL,
	//syscall.SIGIO, 通常は無視
	syscall.SIGIOT,
	syscall.SIGINT,
	syscall.SIGKILL,
	syscall.SIGPIPE,
	syscall.SIGPOLL,
	syscall.SIGPROF,
	syscall.SIGPWR,
	syscall.SIGQUIT,
	syscall.SIGSEGV,
	syscall.SIGSTKFLT,
	syscall.SIGSTOP,
	syscall.SIGSYS,
	syscall.SIGTERM,
	syscall.SIGTSTP,
	syscall.SIGTTIN,
	syscall.SIGTTOU,
	syscall.SIGUNUSED,
	//syscall.SIGURG, 通常は無視
	syscall.SIGTRAP,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
	syscall.SIGVTALRM,
	//syscall.SIGWINCH, 通常は無視
	syscall.SIGXCPU,
	syscall.SIGXFSZ,
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, signals...)
	go func() {
		s := <-sigs
		log.Fatalf("got signal:%s\n", s)
		panic(-1)
	}()

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
