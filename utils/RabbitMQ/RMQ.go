package RMQUtils


type RMQ struct {
	RmqPool *RMQChannelPool
}

func NewRMQ(url string, maxConn int, minChan int, overChan int) (*RMQ, error) {
	connp, e1 := NewRMQConnPool(url, maxConn)
	if e1 != nil {
		return nil, e1
	}
	chanp, e2 := NewRMQChannelPool(connp, minChan, overChan)
	if e2 != nil {
		return nil, e2
	}
	return &RMQ{
		RmqPool: chanp,
	}, nil
}

func (r *RMQ) Get() (*ChannelWithConfirm, error) {
	return r.RmqPool.Get()
}

func (r *RMQ) Put(ch *ChannelWithConfirm) {
	r.RmqPool.Put(ch)
}
