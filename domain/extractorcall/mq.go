package extractorcall

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
)

var (
	ErrClosed        = errors.New("manager has been closed")
	ErrQueueNotFound = errors.New("queue not found in rabbit mq")
)

type MQConnectionConfig struct {
	// RabbitMQ分配的用户名称
	User string
	// RabbitMQ用户的密码
	Pwd string
	// RabbitMQ Broker 的ip地址
	Host string
	// RabbitMQ Broker 监听的端口
	Port string
}

func (c *MQConnectionConfig) ToURL() string {
	return "amqp://" + c.User + ":" + c.Pwd + "@" + c.Host + ":" + c.Port + "/"
}

func GenerateTestMQConnectionConfig() MQConnectionConfig {
	return MQConnectionConfig{
		User: "guest",
		Pwd:  "guest",
		Host: "localhost",
		Port: "5672",
	}
}

type rabbitMQManager struct {
	logger   *logrus.Logger
	conn     *amqp.Connection
	queueMap map[string]*amqp.Queue

	listenMapLock sync.Mutex
	listenMap     map[string]chan<- struct{} // queueName -> stopChan

	closer sync.Once
}

func (mq *rabbitMQManager) Close() error {
	var err error = nil
	closeCalled := false

	mq.closer.Do(func() {
		closeCalled = true

		func() {
			mq.listenMapLock.Lock()
			defer mq.listenMapLock.Unlock()

			for queueName, stopChan := range mq.listenMap {
				mq.logger.Infof("stopping listening of queue [%s] for closing the manager", queueName)
				close(stopChan)
			}
		}()

		err = mq.conn.Close()
	})

	if !closeCalled {
		return ErrClosed
	}

	return err
}

func newRabbitMQManager(url string, queueList []string) (*rabbitMQManager, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, utils.WrapError(err, "create connection fail")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, utils.WrapError(err, "create channel fail")
	}
	defer ch.Close()

	queueMap := make(map[string]*amqp.Queue)
	for _, queueName := range queueList {
		q, err := ch.QueueDeclare(
			queueName,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, utils.WrapError(err, fmt.Sprintf("declare queue [%s] fail", queueName))
		}

		queueMap[queueName] = &q
	}

	return &rabbitMQManager{
		logger:    logging.NewLogger(),
		conn:      conn,
		queueMap:  queueMap,
		listenMap: make(map[string]chan<- struct{}),
	}, nil
}

func (mq *rabbitMQManager) SendObjectByJSON(queueName string, obj any) error {
	queue, ok := mq.queueMap[queueName]
	if !ok {
		return ErrQueueNotFound
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return utils.WrapError(err, "json marshal fail")
	}

	ch, err := mq.conn.Channel()
	if err != nil {
		return utils.WrapError(err, "create channel fail")
	}
	defer ch.Close()

	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         jsonBytes,
		})
	return utils.WrapError(err, "publish fail")
}

func (mq *rabbitMQManager) ListenOn(queueName string, callback func(msg *amqp.Delivery) error) error {
	queue, ok := mq.queueMap[queueName]
	if !ok {
		return ErrQueueNotFound
	}

	ch, err := mq.conn.Channel()
	if err != nil {
		return utils.WrapError(err, "create channel fail")
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return utils.WrapError(err, "create delivery-chan fail")
	}

	stopChan := make(chan struct{})

	mq.listenMapLock.Lock()
	defer mq.listenMapLock.Unlock()

	oldStopChan, ok := mq.listenMap[queueName]
	if ok {
		close(oldStopChan)
	}
	mq.listenMap[queueName] = stopChan

	go func(stopChan2 <-chan struct{}, msgChan <-chan amqp.Delivery, queueName2 string, callback2 func(msg *amqp.Delivery) error) {
		logger := logging.NewLogger()
		for {
			select {
			case msg, alive := <-msgChan:
				if !alive {
					logger.Infof("exiting loop for listening queue [%s] due to channel closed", queueName2)
					return
				}

				logger.Debugf("receive data [%#v]", string(msg.Body))

				err := callback2(&msg)
				if err != nil {
					logger.Errorf("invoking callback for queue [%s] fail with err=%s", queueName2, err)
				}
			case <-stopChan2:
				logger.Infof("exiting loop for listening queue [%s] due to Close signal", queueName2)
				return
			}
		}
	}(stopChan, msgs, queueName, callback)

	return nil
}
