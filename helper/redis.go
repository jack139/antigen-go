package helper

import (
	"log"
	"context"
	"github.com/go-redis/redis/v8"
)


const (
	REDIS_SERVER = "127.0.0.1:7480"
	REDIS_PASSWD = "e18ffb7484f4d69c2acb40008471a71c"
	REDIS_QUEUENAME = "yhfacelib-gosearch-queue"
)

var (
	Rdb *redis.Client
	ctx = context.Background()
)


func InitRDB() error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: REDIS_PASSWD,
		DB:       0,  // use default DB
	})

	if _, err := Rdb.Ping(ctx).Result(); err!=nil {
		return err
	}

	return nil
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
	pubsub := rdb.Subscribe(ctx, "myChannel1")
	ch := pubsub.Channel()
	// Close the subscription when we are done.
	defer pubsub.Close()
	log.Println("Subscribed")

	err := rdb.Publish(ctx, "myChannel1", "payload").Err()
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

