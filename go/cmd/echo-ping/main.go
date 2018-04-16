// Command echo-ping provides a ping-like utility for sending test data
// to an echo-vsock server and verifying the results.
package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mdlayher/vsock"
)

var (
	flagVerbose = flag.Bool("v", false, "enable verbose logging to stderr")
)

type configFlags struct {
	count    uint
	interval time.Duration
	pattern  string
	verify   bool

	contextID uint
	port      uint
}

func main() {
	var c configFlags
	flag.UintVar(&c.count, "count", 0, "number or requests to send (0 sends forever)")
	flag.DurationVar(&c.interval, "interval", 1.0*time.Second, "duration to wait between sending each packet")
	flag.StringVar(&c.pattern, "pattern", "", "pattern to send with request (up to 64 bytes, defaults to md5 hash of the request count)")
	flag.BoolVar(&c.verify, "verify", false, "verify that received data matches sent data")

	flag.UintVar(&c.contextID, "c", 0, "context ID of the remote VM socket")
	flag.UintVar(&c.port, "p", 0, "port ID to connect to")

	flag.Parse()
	log.SetOutput(os.Stderr)

	ping(c)
}

// ping dials a server and sends data to it using VM sockets.  The data
// is read back from the server, and statistics printed to stdout. The
// data is optionally verified to match the sent data.
func ping(cfg configFlags) {
	// Dial a remote server and send a stream to that server.
	c, err := vsock.Dial(uint32(cfg.contextID), uint32(cfg.port))
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer c.Close()

	var p func(i uint) []byte
	p = func(i uint) []byte {
		c := make([]byte, 8)
		binary.LittleEndian.PutUint64(c, uint64(i))
		b := md5.Sum(c)
		return b[:]
	}

	if cfg.pattern != "" {
		b, err := hex.DecodeString(cfg.pattern)
		if err != nil {
			log.Println("pattern must be specified as hex digits")
			log.Fatalf("failed to decode pattern: %v", err)
		}
		p = func(i uint) []byte { return b }
		fmt.Printf("PATTERN: %s", cfg.pattern)
	}

	logf("PING %s FROM %s", c.LocalAddr(), c.RemoteAddr())

	buf := make([]byte, 64)
	tick := time.NewTicker(cfg.interval)
	for i := uint(0); cfg.count == 0 || i < cfg.count; i++ {
		n, err := c.Write(p(i))
		if err != nil {
			log.Fatalf("error writing to socket: %v", err)
		}
		n, err = c.Read(buf)
		fmt.Printf("%d bytes from %s: ping_seq=%d\n", n, c.RemoteAddr(), i)
		<-tick.C
	}
}

// logf shows verbose logging if -v is specified, or does nothing
// if it is not.
func logf(format string, a ...interface{}) {
	if !*flagVerbose {
		return
	}

	log.Printf(format, a...)
}
