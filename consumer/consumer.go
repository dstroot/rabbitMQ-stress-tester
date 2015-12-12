// Package consumer reads (consumes) messages from the declared queue.
package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/dstroot/rabbit-mq-stress-tester/logging"
	"github.com/dstroot/rabbit-mq-stress-tester/queue"
	"github.com/streadway/amqp"
)

// Consume consumes messages
func Consume(uri string, doneChan chan bool, i int) {

	logging.INFO.Printf("I am consumer %d", i+1)

	// get a connection to our server
	logging.INFO.Printf("Consumer %d dialing %q", i+1, uri)
	connection, err := amqp.Dial(uri)
	if err != nil {
		logging.FATAL.Printf("Dial: %s", err.Error())
	}
	defer connection.Close()

	// open a channel on our server
	logging.INFO.Printf("Consumer %d got Connection, getting Channel", i+1)
	channel, err := connection.Channel()
	if err != nil {
		logging.FATAL.Printf("Channel: %s", err.Error())
	}
	defer channel.Close()

	// open queue
	q := queue.MakeQueue(channel)

	// consume messages
	msgs, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	for d := range msgs {
		doneChan <- true
		var thisMessage queue.MqMessage
		err := json.Unmarshal(d.Body, &thisMessage)
		if err != nil {
			log.Printf("Error unmarshalling! %s", err)
		}
		logging.INFO.Printf("Message age: %s", time.Since(thisMessage.TimeNow))
	}

}
