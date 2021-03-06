package kafka

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"syscall"

	"gopkg.in/bsm/sarama-cluster.v2"

	"github.com/Shopify/sarama"
)

var terminatedSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGKILL,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGABRT,
}

func execFunction(msg string, funcs ...interface{}) (result []reflect.Value, err error) {
	funcs = append(funcs, msg)
	f := reflect.ValueOf(funcs[0])
	params := funcs[1:]
	fmt.Printf("All params(%d): %#v\n", len(params), params)

	numbOfReceivedParams := len(params)
	numOfFuncInputParams := f.Type().NumIn()
	if numbOfReceivedParams != numOfFuncInputParams {
		fmt.Println("\033[0;31mError: The number of params is not adapted.\033[0m")
		err = errors.New("The number of params is not adapted.")
		return nil, err
	}

	in := make([]reflect.Value, numbOfReceivedParams)
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return result, nil
}

func consumerGroupCallback(clusterConsumer *cluster.Consumer, funcs ...interface{}) {
	if len(funcs) == 0 {
		for msg := range clusterConsumer.Messages() {
			fmt.Printf("consumed with callback: %s/%d/%d\t\t\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Value)
		}
	} else {
		var wg sync.WaitGroup
		for msg := range clusterConsumer.Messages() {
			wg.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer wg.Done()
				fmt.Printf("consumed with callback: %s/%d/%d\t\t\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Value)
				execFunction(string(msg.Value), funcs...)
				clusterConsumer.MarkPartitionOffset(msg.Topic, msg.Partition, msg.Offset+1, "")
			}(msg)
		}
		wg.Wait()
	}
}

func callback(messages chan *sarama.ConsumerMessage, funcs ...interface{}) {
	if len(funcs) == 0 {
		for msg := range messages {
			fmt.Printf("consumed: %s\n", string(msg.Value))
		}
	} else {
		var wg sync.WaitGroup
		for msg := range messages {

			wg.Add(1)

			go func(msg *sarama.ConsumerMessage) {
				defer wg.Done()
				fmt.Printf("consumed with callback: %s\n", string(msg.Value))
				execFunction(string(msg.Value), funcs...)
			}(msg)
		}

		wg.Wait()
	}
}
