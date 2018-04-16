// Command vsock-echo provides an echo service which listens on vsock
// sockets.
package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"

	"github.com/mdlayher/vsock"
)

var (
	flagVerbose = flag.Bool("v", false, "enable verbose logging to stderr")
)

func main() {
	var (
		flagPort = flag.Uint("p", 0, "port ID to listen on (random port by default)")
	)

	flag.Parse()
	log.SetOutput(os.Stderr)

	port := uint32(*flagPort)

	logf("opening listener: %d", port)

	l, err := vsock.Listen(port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	// Show server's address for setting up client flags.
	log.Printf("receive: listening: %s", l.Addr())

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}
		go echo(c)
	}
}

func echo(c net.Conn) {
	defer c.Close()

	logf("client: %s", c.RemoteAddr())

	n, err := io.Copy(c, c)
	if err != nil {
		log.Fatalf("failed to echo data: %v", err)
	}

	logf("echoed %d bytes to client at %s", n, c.RemoteAddr())
}

// logf shows verbose logging if -v is specified, or does nothing
// if it is not.
func logf(format string, a ...interface{}) {
	if !*flagVerbose {
		return
	}

	log.Printf(format, a...)
}
