package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	var (
		nWorkers int
		retry    int
	)
	flag.IntVar(&nWorkers, "n", 1, "number of forked processes")
	flag.IntVar(&retry, "t", 0, "times to retry to connect")
	flag.Parse()
	if nWorkers < 1 {
		log.Print("-n must be bigger than 0")
		os.Exit(1)
	}
	if retry < 0 {
		log.Print("-t must be bigger than or equal to 0")
		os.Exit(1)
	}

	log.SetFlags(log.Lshortfile)

	conns := make([]io.Writer, 0)
	for i := 0; i < nWorkers; i++ {
		var conn net.Conn
		for j := 0; ; j++ {
			var err error
			conn, err = net.Dial("tcp", flag.Arg(0))
			if err != nil {
				if j < retry {
					continue
				}
				log.Print(err)
			}
			break
		}
		go io.Copy(os.Stdout, conn)
		conns = append(conns, conn)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		for _, conn := range conns {
			io.WriteString(conn, line)
		}
	}
}
