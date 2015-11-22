package queue

import (
	"time"

	"github.com/streadway/amqp"
)

// MqMessage ...
type MqMessage struct {
	TimeNow        time.Time
	SequenceNumber int
	Payload        string
}

// MakeQueue ...
func MakeQueue(c *amqp.Channel) amqp.Queue {

	// Queues in the AMQP model are very similar to queues in other message
	// and task-queueing systems: they store messages that are consumed by
	// applications.
	//
	// Before a queue can be used it has to be declared. QueueDeclare declares
	// a queue to hold messages and deliver to consumers. Declaring creates a
	// queue if it doesn't already exist, or ensures that an existing queue
	// matches the same parameters.
	//
	// Durable and Non-Auto-Deleted queues will survive server restarts and
	// remain when there are no remaining consumers or bindings. Persistent
	// publishings will be restored in this queue on server restart.
	// These queues are only able to be bound to durable exchanges.
	//
	// Non-Durable and Auto-Deleted queues will not be redeclared on server
	// restart and will be deleted by the server after a short time when the
	// last consumer is canceled or the last consumer's channel is closed.
	// Queues with this lifetime can also be deleted normally with QueueDelete.
	// These durable queues can only be bound to non-durable exchanges.
	//
	// Non-Durable and Non-Auto-Deleted queues will remain declared as long
	// as the server is running regardless of how many consumers. This lifetime
	// is useful for temporary topologies that may have long delays between
	// consumer activity. These queues can only be bound to non-durable exchanges.
	//
	// Durable and Auto-Deleted queues will be restored on server restart,
	// but without active consumers, will not survive and be removed. This
	// Lifetime is unlikely to be useful.
	//
	// Exclusive queues are only accessible by the connection that declares
	// them and will be deleted when the connection closes. Channels on other
	// connections will receive an error when attempting declare, bind,
	// consume, purge or delete a queue with the same name.
	//
	// When noWait is true, the queue will assume to be declared on the server.
	// A channel exception will arrive if the conditions are met for existing
	// queues or attempting to modify an existing queue from a different
	// connection.
	//
	// When the error return value is not nil, you can assume the queue
	// could not be declared with these parameters and the channel will be closed.

	/**
	 * [QueueDeclare description]
	 * @param name         string [name of the queue]
	 * @param durable      bool   [durable flag]
	 * @param autoDelete   bool   [auto delete flag]
	 * @param exclusive    bool   [exclusive flag]
	 * @param noWait       bool   [no wait flag]
	 * @param args         Table  [additional arguments]
	 */
	q, err2 := c.QueueDeclare("stress-test-exchange", true, false, false, false, nil)
	if err2 != nil {
		panic(err2)
	}
	return q
}
