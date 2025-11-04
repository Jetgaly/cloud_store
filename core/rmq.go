package core

import (
	"cloud_store/global"
	RMQUtils "cloud_store/utils/RabbitMQ"
	amqp "github.com/rabbitmq/amqp091-go"
)

func InitRMQ() {
	r, e1 := RMQUtils.NewRMQ(global.Config.RMQ.Dsn, 9, 180, 100)
	if e1 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e1.Error())
	}
	channelWithConfirm, e2 := r.Get()
	channel := channelWithConfirm.Channel
	defer r.Put(channelWithConfirm)
	if e2 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e2.Error())
	}
	global.RMQ = r

	//配置延迟clean队列
	//在延迟队列过期之后会转发给死信交换机-->死信队列
	//死信交换机+死信队列(工作队列)
	e4 := channel.ExchangeDeclare(
		"cs.clean.timeoutexc",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if e4 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e4.Error())
	}
	_, e5 := channel.QueueDeclare(
		"cs.clean.timeout",
		true,
		false,
		false, //允许多个消费者连接队列
		false,
		nil,
	)
	if e5 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e5.Error())
	}
	e5 = channel.QueueBind(
		"cs.clean.timeout",
		"cs.clean.timeout",
		"cs.clean.timeoutexc",
		false,
		nil,
	)
	if e5 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e5.Error())
	}

	//延迟交换机+延迟队列
	e6 := channel.ExchangeDeclare(
		"cs.clean.delayexc",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if e6 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e6.Error())

	}
	args := amqp.Table{
		"x-dead-letter-exchange":    "cs.clean.timeoutexc", // 死信转发到主交换机
		"x-dead-letter-routing-key": "cs.clean.timeout",    // 死信路由键
		"x-message-ttl":             300000,                // 队列级别TTL: 5分钟 300000ms
	}
	_, e7 := channel.QueueDeclare(
		"cs.clean.delay",
		true,
		false,
		false, //允许多个消费者连接队列
		false,
		args,
	)
	if e7 != nil {
		global.Logger.Fatal("[RMQ]init fail,%s" + e7.Error())
	}
	e7 = channel.QueueBind(
		"cs.clean.delay",
		"cs.clean.delay",
		"cs.clean.delayexc",
		false,
		nil,
	)
	if e7 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + e7.Error())
	}

	//oss队列
	re4 := channel.ExchangeDeclare(
		"cs.oss.exc",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if re4 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + re4.Error())
	}
	_, re5 := channel.QueueDeclare(
		"cs.oss.queue",
		true,
		false,
		false, //允许多个消费者连接队列
		false,
		nil,
	)
	if re5 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + re5.Error())
	}
	re5 = channel.QueueBind(
		"cs.oss.queue",
		"cs.oss.queue",
		"cs.oss.exc",
		false,
		nil,
	)
	if re5 != nil {
		global.Logger.Fatal("[RMQ]init fail,err: " + re5.Error())
	}
}
