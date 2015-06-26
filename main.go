package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"runtime"
)
import "net"

func main() {
	log.SetFlags(log.Lshortfile)

	ln, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	b := newBroadcaster()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("Connected from %s (%d clients)", conn.RemoteAddr(), b.Count()+1)
		b.Add(conn)
		go handleConnection(conn, b)
	}
}

func handleConnection(conn net.Conn, bc *broadcaster) {
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				conn.Close()
				bc.Remove(conn)
				break
			}
			bc.Send(line)
		}
	}()
}

type broadcaster struct {
	outs        []net.Conn
	quit_chan   chan bool
	add_chan    chan net.Conn
	remove_chan chan net.Conn
	send_chan   chan string
}

func newBroadcaster() *broadcaster {
	b := &broadcaster{
		outs:        make([]net.Conn, 0, 10),
		quit_chan:   make(chan bool),
		add_chan:    make(chan net.Conn),
		remove_chan: make(chan net.Conn),
		send_chan:   make(chan string),
	}
	runtime.SetFinalizer(b, func(b *broadcaster) {
		b.Quit()
	})
	go b.work()
	return b
}

func (b *broadcaster) work() {
	for {
		select {
		case <-b.quit_chan:
			break
		case w := <-b.add_chan:
			b.outs = append(b.outs, w)
		case w := <-b.remove_chan:
			a := b.outs
			for i := 0; i < len(a); i++ {
				if a[i] == w {
					log.Printf("Disconnected from %s (%d clients)", a[i].RemoteAddr(), b.Count()-1)
					b.outs = a[:i+copy(a[i:], a[i+1:])]
					break
				}
			}
		case s := <-b.send_chan:
			for _, w := range b.outs {
				io.WriteString(w, s)
			}
		}
	}
}

func (b *broadcaster) Quit() {
	close(b.quit_chan)
}

func (b *broadcaster) Add(w net.Conn) {
	b.add_chan <- w
}

func (b *broadcaster) Remove(w net.Conn) {
	b.remove_chan <- w
}

func (b *broadcaster) Count() int {
	return len(b.outs)
}

func (b *broadcaster) Send(s string) {
	b.send_chan <- s
}
