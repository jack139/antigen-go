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
	tensor, err := makeTensorFromBytes(image, 256, 256, 0.0, 255.0, true)
	if err!=nil {
		return &map[string]interface{}{"code":9003}, err
	}

	//log.Println(tensor.Value())
	//log.Println("locate tensor: ", tensor.Shape())


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

	log.Println("locate result: ", ret)

	// 使用 locate 结果，进行截图
	cropImage, err := cropBox(image, ret[0])
	if err != nil {
		return &map[string]interface{}{"code":9005}, err
	}

	// 填充成正方形
	cropImage = padBox(cropImage)

	// 转换 为 字节流
	cropByte, err := image2bytes(cropImage)
	if err != nil {
		return &map[string]interface{}{"code":9006}, err
	}

	// ----------- detpos 模型 识别 

	// 转换张量
	tensor, err = makeTensorFromBytes(cropByte, 128, 128, 0.0, 1.0, true)
	if err!=nil {
		return &map[string]interface{}{"code":9007}, err
	}

	//log.Println(tensor.Value())
	//log.Println("detpos tensor: ", tensor.Shape())

	// detpos 模型推理
	res, err = mDetpos.Session.Run(
		map[tf.Output]*tf.Tensor{
			mDetpos.Graph.Operation("input_1").Output(0): tensor,
		},
		[]tf.Output{
			mDetpos.Graph.Operation("dense_1/Softmax").Output(0),
		},
		nil,
	)
	if err != nil {
		return &map[string]interface{}{"code":9008}, err
	}

	ret = res[0].Value().([][]float32)

	log.Printf("detpos result: %v", ret)

	// 转换标签，准备返回结果
	r := bestLabel(ret[0])
	r2 := "invaild"
	if r=="non" {
		r = "none"
	}
	if val, ok := resultMap[r]; ok {
		r2 = val
	}

	return &map[string]interface{}{"result":r2, "comment":r}, nil
}
