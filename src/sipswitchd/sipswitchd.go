package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"sip_parser"

	"telnet"
)

var ()

// Initialize package main
func init() {
	log.Printf("sipswitchd init")
}

func parseSIP(msg string) {
	s := sipparser.ParseMsg(msg)
	if s.Error != nil {
		log.Printf("ParseMsg: error: %s", s.Error)
		return
	}
	log.Printf("--")
	log.Printf("SIP message parse OK")

	sl := s.StartLine

	log.Printf("Start: method=%s type=%s proto=%s version=%s", sl.Method, sl.Type, sl.Proto, sl.Version)
	log.Printf("From: %s", s.From.Val)
	log.Printf("To: %s", s.To.Val)
	log.Printf("CallId: %s", s.CallId)
	log.Printf("UserAgent: %s", s.UserAgent)
	log.Printf("Allow: %s", s.Allow)
	log.Printf("Content-Length: %d", s.ContentLengthInt)

	for i, v := range s.Via {
		log.Printf("Via[%d]: Via=%s", i, v.Via)
		log.Printf("Via[%d]: Transport=%s", i, v.Transport)
		log.Printf("Via[%d]: SentBy=%s", i, v.SentBy)
	}

	log.Printf("SIP: [%s]", s.Msg)
}

// https://groups.google.com/forum/#!topic/golang-nuts/Hm8KaV89Jcs
// http://stackoverflow.com/questions/5884154/golang-read-text-file-into-string-array-and-write
func handleConn(conn net.Conn) {
	defer conn.Close()
	/*
		buf := make([]byte, 70000)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Read: %s", err)
			return
		}
		log.Printf("Read: %d bytes", n)
		parseSIP(buf, n)
	*/

	reader := bufio.NewReader(conn)

	buf := make([]byte, 70000)
	var size int = 0

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			log.Printf("Read: EOF")
			break
		}
		if err != nil {
			log.Printf("Read: error: %s", err)
			return
		}
		log.Printf("Read: %d bytes", n)
		size += n
		break
	}

	parseSIP(string(buf[:size]))
}

func handleTCP(listener net.Listener, addr string) {
	log.Printf("serving SIP on TCP %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept: %s, %s", addr, err)
			continue
		}
		go handleConn(conn)
	}
}

func handleUDP(udpConn *net.UDPConn, addr string) {
	log.Printf("serving SIP on UDP %s", addr)

	if _, err := udpConn.File(); err != nil {
		log.Printf("handleUDP: unable to set UDP connection to blocking: %s", err)
	}

	for {
		/*
			rfc3261 18.1.1 Sending Requests
			However implementations MUST be able to handle messages up to the maximum
			datagram packet size.  For UDP, this size is 65,535 bytes, including
			IP and UDP headers.
		*/
		buf := make([]byte, 66536)
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ReadFromUDP: %s", err)
		}

		log.Printf("ReadFromUDP: %d bytes from %s", n, addr)
		parseSIP(string(buf[:n]))
	}
}

func listenTCP(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("net.Listen: %s, %s", addr, err)
	}
	go handleTCP(listener, addr)
}

func listenUDP(addr string) {
	udpAddr, addrErr := net.ResolveUDPAddr("udp", addr)
	if addrErr != nil {
		log.Printf("net.ResolveUDPAddr: %s, %s", addr, addrErr)
		return
	}
	udpConn, udpErr := net.ListenUDP("udp", udpAddr)
	if udpErr != nil {
		log.Printf("net.ListenUDP: %s, %s", addr, udpErr)
		return
	}

	go handleUDP(udpConn, addr)
}

func listenBoth(addr string) {
	/*
		rfc3261 18.2.1 Receiving Requests
		For any port and interface that a server listens on for UDP,
		it MUST listen on that same port and interface for TCP.
	*/
	listenUDP(addr)
	listenTCP(addr)
}

func inputLoop(rd *bufio.Reader) {
	log.Printf("FIXME WRITEME inputLoop")
}

func outputLoop(wr *bufio.Writer) {
	log.Printf("FIXME WRITEME outputLoop")
}

func handleTelnet(conn net.Conn) {
	defer conn.Close()

	rd, wr := bufio.NewReader(conn), bufio.NewWriter(conn)

	//create userOut channel: will send messages to user

	//create cli interpreter: will write to userOut channel when needed

	//go routine loop:
	//	- read from userOut channel and write into wr
	//	- watch quitOutput channel
	go inputLoop(rd)

	//loop:
	//	- read from rd and feed into cli interpreter
	//	- watch idle timeout
	//	- watch quitInput channel
	outputLoop(wr)
}

func listenTelnet(addr string) {
	telnetServer := telnet.Server{Addr: addr, Handler: handleTelnet}

	log.Printf("serving telnet on TCP %s", addr)

	if err := telnetServer.ListenAndServe(); err != nil {
		log.Fatalf("telnet server on address %s: error: %s", addr, err)
	}
}

func main() {
	log.Printf("sipswitchd booting")

	go listenTelnet(":23")

	addr := ":5060"
	listenBoth(addr)

	log.Printf("sipswitchd ready")

	// wait forever
	<-make(chan int)
}
