package golog

import (
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type AMQPHook struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue

	level logrus.Level

	mu sync.RWMutex
}

func NewAMQPHook(opts ...Option) (*AMQPHook, error) {
	var opt Option

	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.AmqpURL == "" {
		opt.AmqpURL = "amqp://guest:guest@localhost:5672/"
	}

	// if opt.Level == 0 {
	opt.Level = logrus.InfoLevel
	// }

	conn, err := amqp.Dial(opt.AmqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Panicf("failed to open a channel: %v", err)
	}

	queue, err := channel.QueueDeclare(
		"push-loghook", // name
		false,          // type
		false,          //  auto-deleted
		false,          // internal
		false,          // noWait
		nil,            // arguments
	)

	hook := &AMQPHook{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}

	// Run Consumer
	go func() {
		var messages <-chan amqp.Delivery
		messages, err = channel.Consume(
			queue.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("hook.channel.Consume(): %v", err)
		}

		hook.consumer(messages)
	}()

	return hook, nil
}

func (hook *AMQPHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *AMQPHook) Fire(entry *logrus.Entry) error {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	line, err := entry.String()
	if err != nil {
		log.Printf("Unable to read entry, %v", err)
		return err
	}

	log.Printf("logger: %s", line)

	go func() {
		err = hook.channel.Publish(
			"",
			hook.queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(entry.Message),
			})
		if err != nil {
			log.Printf("hook.channel.Publish(): %v", err)
		}
	}()

	return nil
}

func (hook *AMQPHook) Close() error {
	if hook.conn != nil {
		return hook.conn.Close()
	}

	return nil
}

func (hook *AMQPHook) consumer(messages <-chan amqp.Delivery) {
	for d := range messages {
		log.Printf("received a message: %s", d.Body)

		// msg := new(MessagePushWebHook)
		// if err := json.Unmarshal(d.Body, &msg); err != nil {
		// 	logger.Errorf(fmt.Sprintf("Invalid Data"))
		// 	continue
		// }
		//
		// if err := hook.pushWebHook(msg); err != nil {
		// 	continue
		// }
	}
}

// func (hook *AMQPHook) pushWebHook(msg *MessagePushWebHook) error {
// 	client := &http.Client{
// 		Timeout: 5 * time.Second,
// 	}
// 	msgPayload, err := json.Marshal(msg.Payload)
// 	if err != nil {
// 		logger.Errorf(fmt.Sprintf("json.Marshal(): %v", err))
// 		return err
// 	}
//
// 	req, err := http.NewRequestWithContext(context.Background(), "POST", msg.UrlWebHook, bytes.NewReader(msgPayload))
// 	if err != nil {
// 		logger.Errorf(fmt.Sprintf("http.NewRequest(): %v", err))
// 		return err
// 	}
//
// 	req.Header.Add("Content-Type", "application/json")
//
// 	res, err := client.Do(req)
// 	if err != nil {
// 		logger.Errorf(fmt.Sprintf("client.Do(): %v", err))
// 		return err
// 	}
//
// 	defer res.Body.Close()
//
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		logger.Errorf(fmt.Sprintf("ioutil.ReadAll(): %v", err))
// 		return err
// 	}
//
// 	fmt.Println(string(body))
// 	return nil
// }
