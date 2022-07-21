package api

import (
	"log"
	//"fmt"
)


/* 空接口 */
func ApiNothing(data *map[string]interface{}) (*map[string]interface{}, error) {
	log.Println("ApiNonthing")

	log.Printf("%v\n", *data)

	return data, nil // 正确返回
	//return &map[string]interface{}{"code":9003}, fmt.Errorf("error test") // 错误返回： 错误代码，错误信息
}
