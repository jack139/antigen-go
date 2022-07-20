package helper

import (
	"log"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)


const (
	REDIS_SERVER = "127.0.0.1:7480"
	REDIS_PASSWD = "e18ffb7484f4d69c2acb40008471a71c"
	REDIS_QUEUENAME = "goinfer-synchronous-asynchronous-queue"
)

var (
	Rdb *redis.Client
	//ctx = context.Background()
)

func InitRDB() error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: REDIS_PASSWD,
		DB:       0,  // use default DB
	})

	if _, err := Rdb.Ping(context.Background()).Result(); err!=nil {
		return err
	}

	log.Println("Redis connected.")

	return nil
}

func Redis_publish(queue string, message string) error {
	if queue=="NO_RECIEVER" {
		return nil
	}

	err := Rdb.Publish(context.Background(), queue, message).Err()
	if err != nil {
		return err
	}

	return nil
}


func Redis_publish_request(requestId string, data *map[string]interface{}) error {
	msgBodyMap := map[string]interface{}{
		"request_id": requestId,
		"data": *data,
	}
	msgBody, err := json.Marshal(msgBodyMap)
	if err != nil {
		return err
	}

	queue := REDIS_QUEUENAME // todo: 多队列处理

	log.Println(queue, msgBodyMap)

	return Redis_publish(queue, string(msgBody))
}


func redisTest(){

	log.Println("start")

	rdb := redis.NewClient(&redis.Options{
		Addr:	  "127.0.0.1:7480",
		Password: "e18ffb7484f4d69c2acb40008471a71c", 
		DB:		  0,  // use default DB
	})

	log.Println("rdb")

	// There is no error because go-redis automatically reconnects on error.
	pubsub := rdb.Subscribe(context.Background(), REDIS_QUEUENAME)
	ch := pubsub.Channel()
	// Close the subscription when we are done.
	defer pubsub.Close()
	log.Println("Subscribed")

	err := rdb.Publish(context.Background(), REDIS_QUEUENAME, "payload").Err()
	if err != nil {
		panic(err)
	}

	log.Println("published")

	for msg := range ch {
		log.Println(msg.Channel, msg.Payload)
		break
	}

	log.Println("left")
}


func init(){
	//redisTest()
}
