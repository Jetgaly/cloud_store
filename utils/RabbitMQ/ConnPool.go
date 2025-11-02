package RMQUtils

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RMQConnPool struct {
	conns   chan *amqp.Connection
	url     string
	maxConn int
}

func NewRMQConnPool(url string, maxConn int) (*RMQConnPool, error) {
	pool := &RMQConnPool{
		conns:   make(chan *amqp.Connection, maxConn),
		url:     url,
		maxConn: maxConn,
	}

	for i := 0; i < maxConn; i++ {
		conn, err := amqp.Dial(url)
		if err != nil {
			return nil, err
		}

		pool.conns <- conn
	}

	return pool, nil
}

func (p *RMQConnPool) Get() (*amqp.Connection, error) {
	select {
	case conn, ok := <-p.conns:
		if ok {
			if conn.IsClosed() {
				var err error
				conn, err = amqp.Dial(p.url)
				if err != nil {
					return nil, err
				}
			}
			return conn, nil
		} else {
			return nil, ErrConnCantTake
		}
	case <-time.After(3 * time.Second):
		return nil, ErrTimeout
	}
}

func (p *RMQConnPool) Put(conn *amqp.Connection) {
	select {
	case p.conns <- conn:
		return
	default:
		conn.Close()
	}
}
