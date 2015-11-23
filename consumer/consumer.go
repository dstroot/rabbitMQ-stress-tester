// Package consumer reads (consumes) messages from the declared queue.
package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/dstroot/rabbit-mq-stress-tester/queue"
	"github.com/streadway/amqp"
)

// Consume cosumes messages
func Consume(uri string, doneChan chan bool) {

	// create connection
	connection, err := amqp.Dial(uri)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer connection.Close()

	// create channel
	channel, err1 := connection.Channel()
	if err1 != nil {
		log.Fatalf("Error: %s", err1)
	}
	defer channel.Close()

	// open queue
	q := queue.MakeQueue(channel)

	// consume messages
	msgs, err2 := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err2 != nil {
		log.Fatalf("Error: %s", err2)
	}

	for d := range msgs {
		doneChan <- true
		var thisMessage queue.MqMessage
		err3 := json.Unmarshal(d.Body, &thisMessage)
		if err3 != nil {
			log.Printf("Error unmarshalling! %s", err3)
		}
		log.Printf("Message age: %s", time.Since(thisMessage.TimeNow))
	}

	log.Println("done recieving")
}
