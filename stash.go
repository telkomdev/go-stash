package stash

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	// BrokenPipeError const type to check whether the write process to tcp is broken
	BrokenPipeError = "broken pipe"
)

// CRLF (Carriage Return and Line Feed in ASCII code)
var CRLF = []byte{13, 10}

func addCRLF(data []byte) []byte {
	return append(data, CRLF...)
}

// Stash structure
type Stash struct {
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	o            *options
	address      string
}

// Option function
type Option func(*options)

type options struct {
	dialer       *net.Dialer
	protocol     string
	useTLS       bool
	skipVerify   bool
	readTimeout  time.Duration
	writeTimeout time.Duration
	tlsConfig    *tls.Config
}

// SetTLS Option func
func SetTLS(useTLS bool) Option {
	return func(o *options) {
		o.useTLS = useTLS
	}
}

// SetSkipVerify Option func
func SetSkipVerify(skipVerify bool) Option {
	return func(o *options) {
		o.skipVerify = skipVerify
	}
}

// SetReadTimeout Option func
func SetReadTimeout(readTimeout time.Duration) Option {
	return func(o *options) {
		o.readTimeout = readTimeout
	}
}

// SetWriteTimeout Option func
func SetWriteTimeout(writeTimeout time.Duration) Option {
	return func(o *options) {
		o.writeTimeout = writeTimeout
	}
}

// SetKeepAlive Option func
func SetKeepAlive(keepAlive time.Duration) Option {
	return func(o *options) {
		o.dialer.KeepAlive = keepAlive
	}
}

// SetTLSConfig Option func
func SetTLSConfig(config *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = config
	}
}

// SetProtocolConn Option func
// set protocol connection between logstash : `tcp` or `udp`
func SetProtocolConn(protocol string) Option {
	return func(o *options) {
		o.protocol = protocol
	}
}

func (s *Stash) dial(address string, o *options) error {
	conn, err := net.Dial(o.protocol, address)
	if err != nil {
		return err
	}

	s.conn = conn

	// if useTLS true
	// Force stash to use TLS
	if o.useTLS {
		var tlsConfig *tls.Config
		if o.tlsConfig == nil {
			tlsConfig = &tls.Config{InsecureSkipVerify: o.skipVerify}
		} else {
			tlsConfig = o.tlsConfig
		}

		if tlsConfig.ServerName == "" {
			host, _, err := net.SplitHostPort(s.address)
			if err != nil {
				conn.Close()
				return err
			}
			tlsConfig.ServerName = host
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return err
		}

		// replace current Conn object with tlsConn
		s.conn = tlsConn
	}

	return nil
}

// Connect function, this function will connect to logstash server
func Connect(host string, port uint64, opts ...Option) (*Stash, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	s := &Stash{address: address}

	o := &options{
		dialer: &net.Dialer{
			KeepAlive: time.Minute * 5,
		},
		protocol:     "tcp",
		writeTimeout: 30 * time.Second,
		readTimeout:  30 * time.Second,
	}
	for _, option := range opts {
		option(o)
	}

	s.o = o
	if err := s.dial(address, o); err != nil {
		return nil, err
	}

	s.readTimeout = o.readTimeout
	s.writeTimeout = o.writeTimeout
	return s, nil
}

// Write function, implement from io.Writer
func (s *Stash) Write(data []byte) (int, error) {
	if s.writeTimeout != 0 {
		deadline := time.Now().Add(s.writeTimeout)
		s.conn.SetWriteDeadline(deadline)
	}

	// remove any Carriage Return or Line Feed in bytes data
	// before concate with new Carriage Return and Line Feed
	data = bytes.Trim(data, string(CRLF))

	// concate with new Carriage Return or Line Feed
	data = addCRLF(data)

	// write data to Connection
	_, err := s.conn.Write(data)
	if err != nil {
		if strings.Contains(err.Error(), BrokenPipeError) {
			log.Printf("go-stash: %s | do re dial\n", err.Error())
			// re dial ignore error
			err = s.dial(s.address, s.o)
			if err != nil {
				log.Printf("go-stash: %s | do re dial\n", err.Error())
			}
		} else {
			return 0, err
		}
	}
	return len(data), nil
}

// Close function, will close connection
func (s *Stash) Close() error {
	return s.conn.Close()
}
