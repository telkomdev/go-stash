package stash

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

var (
	CRLF = []byte{13, 10}
)

func addCRLF(data []byte) []byte {
	return append(data, CRLF...)
}

type Stash struct {
	conn         net.Conn
	bw           *bufio.Writer
	br           *bufio.Reader
	readTimeout  time.Duration
	writeTimeout time.Duration
	address      string
}

// Option function
type Option func(*options)

type options struct {
	dialer       *net.Dialer
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
		o.readTimeout = writeTimeout
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

func Connect(host string, port uint64, opts ...Option) (*Stash, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	s := &Stash{address: address}

	o := &options{
		dialer: &net.Dialer{
			KeepAlive: time.Minute * 5,
		},
	}
	for _, option := range opts {
		option(o)
	}

	if o.readTimeout == 0 {
		o.readTimeout = 5
	}

	if o.writeTimeout == 0 {
		o.writeTimeout = 5
	}

	conn, err := o.dialer.Dial("tcp", s.address)

	if err != nil {
		return nil, err
	}

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
				return nil, err
			}
			tlsConfig.ServerName = host
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}
		conn = tlsConn
	}

	s.conn = conn
	s.bw = bufio.NewWriter(s.conn)
	s.br = bufio.NewReader(s.conn)
	s.readTimeout = o.readTimeout
	s.writeTimeout = o.writeTimeout
	return s, nil
}

// Write function, implement from io.Writer
func (s *Stash) Write(data []byte) (int, error) {
	if s.writeTimeout != 0 {
		deadline := time.Now().Add(s.writeTimeout * time.Millisecond)
		s.conn.SetWriteDeadline(deadline)
	}

	data = addCRLF(data)
	_, err := s.bw.Write(data)
	if err != nil {
		return 0, err
	}
	s.bw.Flush()
	return len(data), nil
}

// Close function, will close connection
func (s *Stash) Close() error {
	return s.conn.Close()
}
