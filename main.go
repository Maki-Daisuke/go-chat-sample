package main

import (
	"bufio"
	"io"
	"log"
	"os"
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
		log.Print("Connected from ", conn.RemoteAddr())
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
				log.Print("Disconnected from ", conn.RemoteAddr())
				conn.Close()
				bc.Remove(conn)
				break
			}
			bc.Send(line)
		}
	}()
}

type broadcaster struct {
	outs        []io.Writer
	quit_chan   chan bool
	add_chan    chan io.Writer
	remove_chan chan io.Writer
	send_chan   chan string
}

func newBroadcaster() *broadcaster {
	b := &broadcaster{
		outs:        make([]io.Writer, 0, 10),
		quit_chan:   make(chan bool),
		add_chan:    make(chan io.Writer),
		remove_chan: make(chan io.Writer),
		send_chan:   make(chan string),
	}
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

func (b *broadcaster) Add(w io.Writer) {
	b.add_chan <- w
}

func (b *broadcaster) Remove(w io.Writer) {
	b.remove_chan <- w
}

func (b *broadcaster) Send(s string) {
	b.send_chan <- s
}
