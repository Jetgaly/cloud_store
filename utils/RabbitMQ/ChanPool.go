package RMQUtils

import (
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrTimeout      error = errors.New("timeout")
	ErrConnCantTake error = errors.New("rmq can't take conn")
)

type ChannelWithConfirm struct {
	Channel *amqp.Channel
	Confirm *chan amqp.Confirmation
}

type RMQChannelPool struct {
	channels  chan *ChannelWithConfirm
	Pool      *RMQConnPool
	minChan   int
	overChan  int
	mutex     sync.Mutex
	overCount int
}

func NewRMQChannelPool(connPool *RMQConnPool, minChan int, overChan int) (*RMQChannelPool, error) {
	channelPool := &RMQChannelPool{
		channels: make(chan *ChannelWithConfirm, minChan),
		Pool:     connPool,
		minChan:  minChan,
		overChan: overChan,
	}

	for i := 0; i < minChan; i++ {
		conn, err := connPool.Get()
		if err != nil {
			return nil, err
		}

		ch, err := conn.Channel()
		if err != nil {
			connPool.Put(conn)
			return nil, err
		}
		err = ch.Confirm(false)
		if err != nil {
			connPool.Put(conn)
			return nil, err
		}

		confirm := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
		//x := <-confirm
		//x.DeliveryTag 批量发送时使用

		channelPool.channels <- &ChannelWithConfirm{
			Channel: ch,
			Confirm: &confirm,
		}
		connPool.Put(conn)
	}

	return channelPool, nil
}

func (p *RMQChannelPool) Get() (*ChannelWithConfirm, error) {
	select {
	case ch := <-p.channels:
		if ch.Channel.IsClosed() {
			conn, err := p.Pool.Get()
			if err != nil {
				return nil, err
			}
			var channel *amqp.Channel
			channel, err = conn.Channel()
			if err != nil {
				p.Pool.Put(conn)
				return nil, err
			}
			err = channel.Confirm(false)
			if err != nil {
				p.Pool.Put(conn)
				return nil, err
			}
			confirm := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
			ch = &ChannelWithConfirm{
				Channel: channel,
				Confirm: &confirm,
			}
			p.Pool.Put(conn)
		}
		return ch, nil
	default:
		p.mutex.Lock()
		if p.overCount < p.overChan {
			conn, err := p.Pool.Get()
			if err != nil {
				return nil, err
			}
			c, e1 := conn.Channel()
			p.Pool.Put(conn)
			p.overCount++
			p.mutex.Unlock()
			e2 := c.Confirm(false)
			if e1 != nil || e2 != nil {
				return nil, errors.New("[RMQ]create channelWithConfirm err")
			}
			confirm := c.NotifyPublish(make(chan amqp.Confirmation, 1))
			return &ChannelWithConfirm{
				Channel: c,
				Confirm: &confirm,
			}, e1
		} else {
			select {
			case ch := <-p.channels:
				return ch, nil
			case <-time.After(3 * time.Second):
				return nil, ErrTimeout
			}
		}
	}
}

func (p *RMQChannelPool) Put(ch *ChannelWithConfirm) {
	select {
	case p.channels <- ch:
	default:
		p.mutex.Lock()
		p.overCount--
		p.mutex.Unlock()
		ch.Channel.Close()
	}
}
