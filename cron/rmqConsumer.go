package cron

import (
	"cloud_store/global"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Consumers struct {
	Conn    *amqp091.Connection
	Queue   string
	Handler func(msg []byte) error
	Count   int //消费者数量
	Wg      sync.WaitGroup
	Cancel  context.CancelFunc
}

// Start 启动多个消费者协程
func (c *Consumers) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.Cancel = cancel

	for i := 0; i < c.Count; i++ {
		c.Wg.Add(1)
		go c.worker(ctx, i)
	}

	global.Logger.Info("启动" + strconv.Itoa(c.Count) + "个消费者协程，队列: " + c.Queue)
	return nil
}

// NewConsumer 实例化一个消费者, 会单独用一个channel，设置每次只取一个消息
func (c *Consumers) worker(ctx context.Context, no int) {
	defer c.Wg.Done()

	logMsg := fmt.Sprintf("worker:%d starts", no)
	global.Logger.Info(logMsg)

	err := c.consumeMessage(ctx, no)
	if err != nil {
		global.Logger.Fatal(fmt.Sprintf("worker-%d err: %s", no, err.Error()))
	}

}
func (c *Consumers) consumeMessage(ctx context.Context, no int) error {

	ch, err := c.Conn.Channel()
	if err != nil {
		return fmt.Errorf("new mq channel err: %s", err.Error())
	}
	defer ch.Close()
	// 设置 QoS：每次只取一个消息处理
	// 参数说明：
	// prefetchCount: 每次预取的消息数量，设为1表示每次只取一个
	// prefetchSize: 预取的消息总大小（字节），0表示不限制
	// global: 是否在连接级别应用，false表示只在当前channel生效
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("set qos err: %v", err)
	}

	deliveries, err := ch.Consume(c.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume err: %v, queue: %s", err, c.Queue)
	}

	for {
		select {
		case <-ctx.Done():
			global.Logger.Info(fmt.Sprintf("worker:%d stop", no))
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				global.Logger.Error("msg queue has closed")
			} else {
				err = c.Handler(delivery.Body)
				if err != nil {
					_ = delivery.Reject(true) // 处理失败，重新入队
					//logc
					global.Logger.Error(fmt.Sprintf("msg handler err:%s", err.Error()))

				} else {
					_ = delivery.Ack(false) // 处理成功，确认消息
				}
			}
		}
	}
}

// Stop 优雅停止消费者
func (c *Consumers) Stop() {
	defer c.Conn.Close()
	global.Logger.Info("stopping consumers gracefully")

	if c.Cancel != nil {
		c.Cancel() // 发送关闭信号
	}

	// 等待所有消费者协程完成
	done := make(chan struct{})
	go func() {
		c.Wg.Wait()
		close(done)
	}()

	// 设置超时
	select {
	case <-done:
		global.Logger.Info(fmt.Sprintf("consumers:%s have stopped", c.Queue))
	case <-time.After(30 * time.Second):
		global.Logger.Info("consumers stopping timeout 30 sec, quit forcefuly")
	}

}
