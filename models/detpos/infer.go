package detpos

import (
	"log"
	"fmt"
	"io/ioutil"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/jack139/go-infer/helper"
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

func modleInfer(image []byte) (string, int, error){

	// 转换张量
	tensor, err := makeTensorFromBytes(image, 256, 256, 0.0, 255.0, true)
	if err!=nil {
		return "", 9003, err
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
		return "", 9004, err
	}

	ret := res[0].Value().([][]float32)

	log.Println("locate result: ", ret)

	// 使用 locate 结果，进行截图
	cropImage, err := cropBox(image, ret[0])
	if err != nil {
		return "", 9005, err
	}

	if cropImage == nil { // 未定位到 目标， 返回 none 结果
		return "none", 0, nil
	}

	// 填充成正方形
	cropImage = padBox(cropImage)

	// 转换 为 字节流
	cropByte, err := image2bytes(cropImage)
	if err != nil {
		return "", 9006, err
	}

	// ----------- detpos 模型 识别 

	// 转换张量
	tensor, err = makeTensorFromBytes(cropByte, 128, 128, 0.0, 1.0, true)
	if err!=nil {
		return "", 9007, err
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
		return "", 9008, err
	}

	ret = res[0].Value().([][]float32)

	log.Printf("detpos result: %v", ret)

	// 转换标签，准备返回结果
	r := bestLabel(ret[0])

	return r, 0, nil
}

// 模型热身
func warmup(path string){
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("warmup fail: %s", err.Error())
		return
	}

	for _, file := range files {
		if file.IsDir() { continue }
	
		image, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, file.Name()))
		if err != nil { continue }

		r, _, err := modleInfer(image)
		if err==nil {
			log.Printf("warmup: %s %s", file.Name(), r)
		} else {
			log.Printf("warmup fail: %s %s", file.Name(), err.Error())
		}
	}
}
