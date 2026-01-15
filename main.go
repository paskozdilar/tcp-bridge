package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	host  string
	ports []string
	pairs []portPair
	n     atomic.Int32
)

type portPair struct {
	src string
	dst string
}

var usage string = fmt.Sprintf(
	strings.Join([]string{
		"%s: [-h|--help] HOST PORTS...",
		"",
		"Forwards all TCP connections to HOST, from a list of PORTS.",
		"",
		"Arguments:",
		"    HOST    destination host or address (e.g. \"example.com\")",
		"    PORTS   space-delimited list of ports (e.g. \"80 443\")",
		"            or pairs of port mappings (e.g. \"80->8080 443->8443\")",
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
	for _, pair := range pairs {
		wg.Add(1)
		go func(pair portPair) {
			defer wg.Done()
			forwardPortToPort(pair)
		}(pair)
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
		pattern := "[0-9]+(-[0-9]+)?"
		valid, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("internal error: invalid pattern %s", pattern)
		}
		for _, arg := range os.Args[2:] {
			if !valid.MatchString(arg) {
				return fmt.Errorf("invalid argument: %s", arg)
			}
			port, err := strconv.Atoi(arg)
			if err == nil {
				if port < 0 || port > 65535 {
					return fmt.Errorf("argument '%s': not in range [1,65535]", arg)
				}
				ports = append(ports, arg)
				continue
			}
			arr := strings.SplitN(arg, "-", 2)
			pair := portPair{src: arr[0], dst: arr[1]}
			pairs = append(pairs, pair)
		}
	}
	return nil
}

func forwardPort(port string) {
	log := log.New(log.Writer(),
		fmt.Sprintf("PORT %s [%d]: ", port, n.Add(1)),
		log.Flags())

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	log.Printf("forwarding port %[1]s to %[2]s:%[1]s", port, host)

	for {
		src, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		dst, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
		if err != nil {
			log.Printf("dial error to %s:%s: %v", host, port, err)
			src.Close()
			continue
		}

		go func() {
			io.Copy(src, dst)
			src.Close()
			dst.Close()
		}()

		go func() {
			io.Copy(src, dst)
			src.Close()
			dst.Close()
		}()
	}
}

func forwardPortToPort(pair portPair) {
	log := log.New(log.Writer(),
		fmt.Sprintf("PORT %s-%s [%d]: ", pair.src, pair.dst, n.Add(1)),
		log.Flags())

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", pair.src))
	if err != nil {
		log.Printf("error listening on port %s: %v", pair.src, err)
		return
	}
	defer l.Close()

	log.Printf("forwarding port %s to %s:%s", pair.src, host, pair.dst)

	for {
		src, err := l.Accept()
		if err != nil {
			log.Printf("accept error on %s: %v", pair.src, err)
			return
		}

		dst, err := net.DialTimeout("tcp", net.JoinHostPort(host, pair.dst), 5*time.Second)
		if err != nil {
			log.Printf("dial error to %s:%s: %v", host, pair.dst, err)
			src.Close()
			continue
		}

		go func(s, d net.Conn) {
			defer s.Close()
			defer d.Close()

			errChan := make(chan error, 2)
			cp := func(dst io.Writer, src io.Reader) {
				_, err := io.Copy(dst, src)
				errChan <- err
			}

			go cp(s, d)
			go cp(d, s)

			// Wait for one side to finish or error
			<-errChan
		}(src, dst)
	}
}
