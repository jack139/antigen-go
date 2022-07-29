package detpos

import (
	"fmt"
	"log"
	"strconv"
	"encoding/base64"

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
func (x *DetPos) Infer(reqData *map[string]interface{}) (*map[string]interface{}, error) {
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

	// 转换张量
	tensor, err := makeTensorFromBytes(image)
	if err!=nil {
		return &map[string]interface{}{"code":9003}, err
	}

	log.Println(tensor.Value())
	log.Println(tensor.Shape())

	// locate 模型推理
	res, err := mLocate.Session.Run(
		map[tf.Output]*tf.Tensor{
			mLocate.Graph.Operation("input_1").Output(0): tensor,
		},
		[]tf.Output{
			mLocate.Graph.Operation("dense_3/Sigmoid").Output(0),
		},
		nil,
	)
	if err != nil {
		return &map[string]interface{}{"code":9004}, err
	}

	ret := res[0].Value().([][]float32)

	log.Println(ret)

	box, rotateAngle := cropBox(image, ret[0])

	log.Println(box, rotateAngle)

	return &map[string]interface{}{"embeddings":ret[0]}, nil

}
