package detpos

import (
	"os"
	"time"
	"fmt"
	"log"
	"strconv"
	"encoding/base64"
	"io/ioutil"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/jack139/go-infer/helper"
)

const (
	MaxSeqLength = 512
)

/* 训练好的模型权重 */
var (
	mLocate *tf.SavedModel
	mDetpos *tf.SavedModel
)

/* 初始化模型 */
func initModel() error {
	var err error
	mLocate, err = tf.LoadSavedModel(helper.Settings.Customer["LocateModelPath"], []string{"train"}, nil)
	if err != nil {
		return err
	}

	mDetpos, err = tf.LoadSavedModel(helper.Settings.Customer["DetposModelPath"], []string{"train"}, nil)
	if err != nil {
		return err
	}


	// 模型热身
	warmup(helper.Settings.Customer["WARM_UP_IMAGES"])

	return nil
}


/*  定义模型相关参数和方法  */
type DetPos struct{}

func (x *DetPos) Init() error {
	return initModel()
}

func (x *DetPos) ApiPath() string {
	return "/antigen/check"
}

func (x *DetPos) ApiEntry(reqData *map[string]interface{}) (*map[string]interface{}, error) {
	log.Println("Api_DetPos")

	// 检查参数
	imageBase64, ok := (*reqData)["image"].(string)
	if !ok {
		return &map[string]interface{}{"code":9001}, fmt.Errorf("need image")
	}

	// 构建请求参数
	reqDataMap := map[string]interface{}{
		"image": imageBase64,
	}

	return &reqDataMap, nil
}


// 推理
func (x *DetPos) Infer(requestId string, reqData *map[string]interface{}) (*map[string]interface{}, error) {
	log.Println("Infer_DetPos")

	imageBase64 := (*reqData)["image"].(string)

	// 解码base64
	image, err  := base64.StdEncoding.DecodeString(imageBase64)
	if err!=nil {
		return &map[string]interface{}{"code":9901}, err
	}

	// 检查图片大小
	maxSize, _ := strconv.Atoi(helper.Settings.Customer["MAX_IMAGE_SIZE"])
	if len(image) > maxSize {
		return &map[string]interface{}{"code":9002}, fmt.Errorf("图片数据太大")
	}

	// 模型推理
	r, code, err := modleInfer(image)
	if err != nil {
		return &map[string]interface{}{"code":code}, err
	}

	r2 := "invaild"
	if r == "non" {
		r = "none"
	}
	if val, ok := resultMap[r]; ok {
		r2 = val
	}

	// 保存请求图片和识别结果（文件名中体现结果）
	if r != "neg" { // neg 结果不保存
		if helper.Settings.Customer["SAVE_IMAGE"] == "1" {
			output_dir := fmt.Sprintf("%s/%s", 
				helper.Settings.Customer["SAVE_IMAGE_PATH"], 
				time.Now().Format("20060102"))
			err = os.Mkdir(output_dir, 0755) // 建日志目录， 日期 做子目录
			if err == nil || os.IsExist(err) { // 不处理错误
				_ = ioutil.WriteFile(fmt.Sprintf("%s/%s_%s.jpg", output_dir, requestId, r), image, 0644)
			} else {
				log.Println("ERROR when saving log: ", err.Error())
			}
		}
	}

	return &map[string]interface{}{"result":r2, "comment":r}, nil
}
