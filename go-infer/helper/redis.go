package helper

import (
	"fmt"
	"log"
	"time"
	"context"
	"strconv"
	"encoding/json"
	"math/rand"
	"crypto/md5"
	"github.com/go-redis/redis/v8"
)


const (
	REDIS_SERVER = "127.0.0.1:7480"
	REDIS_PASSWD = "e18ffb7484f4d69c2acb40008471a71c"
	REDIS_QUEUENAME = "goinfer-synchronous-asynchronous-queue"
	REQUEST_QUEUE_NUM = 1 // 队列数量
	MESSAGE_TIMEOUT = 10 // 超时时间
)

var (
	Rdb *redis.Client

	/* 随即字符串的字母表 */
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init(){

	// 初始化随机数发生器
	rand.Seed(time.Now().UnixNano())

	// 初始化redis连接
	err := redis_init()
	if err!=nil {
		log.Fatal("Redis connecting FAIL: ", err)
	}

	//redisTest()
}


/* 产生随机串 */
func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

/* 产生 request id */
func GenerateRequestId() string {
	year, month, day := time.Now().Date()
	h := md5.New()
	h.Write([]byte(randSeq(10)))
	sum := h.Sum(nil)
	md5Str := fmt.Sprintf("%4d%02d%02d%x", year, month, day, sum)
	return md5Str
}


func redis_init() error {
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

// 发布消息
func Redis_publish(queue string, message string) error {
	if queue=="NO_RECIEVER" {
		return nil
	}

	err := Rdb.Publish(context.Background(), queue, message).Err()
	if err != nil {
		return err
	}

	log.Printf("--> %s [%d]", queue, len(message))

	return nil
}

/* 返回随机队列号码 */
func choose_queue_random() string {
	return strconv.Itoa(rand.Intn(REQUEST_QUEUE_NUM))
}

// 发布 请求数据 到 处理队列
func Redis_publish_request(requestId string, data *map[string]interface{}) error {
	msgBodyMap := map[string]interface{}{
		"request_id": requestId,
		"data": *data,
	}
	msgBody, err := json.Marshal(msgBodyMap)
	if err != nil {
		return err
	}

	queue := REDIS_QUEUENAME + choose_queue_random() // 多队列处理

	//log.Println(queue, msgBodyMap)

	return Redis_publish(queue, string(msgBody))
}


// 订阅消息
func Redis_subscribe(requestId string) *redis.PubSub {
	return Rdb.Subscribe(context.Background(), requestId)
}

// 接受订阅的消息，只收一条
func Redis_sub_receive(pubsub *redis.PubSub) (retBytes []byte) {
	startTime := time.Now().Unix()
	for {
		msgi, err := pubsub.ReceiveTimeout(context.Background(), time.Millisecond)
		if err == nil {
			if msg, ok := msgi.(*redis.Message); ok {
				log.Printf("<-- %s [%d]", msg.Channel, len(msg.Payload))
				//log.Println("output: ", msg.Payload)
				retBytes = []byte(msg.Payload)
				break
			}
		}

		// 检查超时
		if time.Now().Unix() - startTime > MESSAGE_TIMEOUT {
			retBytes = []byte("{\"code\":9997,\"msg\":\"消息队列超时\"}")
			break
		}

		time.Sleep(2 * time.Millisecond)
	}

	return retBytes
}


/*----------------------------------------------------------*/

func redisTest(){

	log.Println("start")

	rdb := redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: REDIS_PASSWD,
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

