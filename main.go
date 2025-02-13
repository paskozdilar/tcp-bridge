package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	host  string
	ports []string
)

var usage string = fmt.Sprintf(
	strings.Join([]string{
		"%s: [-h|--help] HOST PORTS...",
		"",
		"Forwards all TCP connections to HOST, from a list of PORTS.",
		"",
		"Arguments:",
		"    HOST    destination host or address (e.g. \"example.com\")",
		"    PORTS   space-delimited list of ports (e.g. \"80 443\")",
	}, "\n"),
	os.Args[0],
)

func main() {
	log.SetFlags(0)
	if err := parse(); err != nil {
		log.Printf("parse error: %s\nuse --help to print usage", err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, port := range ports {
		wg.Add(1)
		go func(port string) {
			defer wg.Done()
			forwardPort(port)
		}(port)
	}
}

func parse() error {
	if len(os.Args) == 1 || (len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help")) {
		fmt.Fprintln(os.Stdout, usage)
		os.Exit(1)
	} else if len(os.Args) < 3 {
		return errors.New("too few arguments")
	} else {
		host = os.Args[1]
		for _, arg := range os.Args[2:] {
			port, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("argument '%s': %w", arg, err)
			}
			if port < 0 || port > 65535 {
				return fmt.Errorf("argument '%s': not in range [1,65535]", arg)
			}
			ports = append(ports, arg)
		}
	}
	return nil
}

func forwardPort(port string) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	log.Printf("forwarding port %s to %s:%s", port, host, port)

	for {
		src, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		dst, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
		if err != nil {
			log.Println(err)
			src.Close()
			continue
		}

		go func() {
			io.Copy(src, dst)
			src.Close()
			dst.Close()
		}()

		io.Copy(src, dst)
		src.Close()
		dst.Close()
	}
}
