package detpos

import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	"bytes"
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

var (
	detposLabels = []string{"fal", "neg", "non", "nul", "pos"}

	resultMap = map[string]string{
		"pos"  : "positive",
		"neg"  : "negative",
		"none" : "invalid", //'not_found',
		"fal"  : "invalid",
		"nul"  : "invalid",
	}
)

/*
	below codes taken from
	https://github.com/tensorflow/tensorflow/blob/master/tensorflow/go/example_inception_inference_test.go
*/

// This function constructs a graph of TensorFlow operations which takes as
// input a JPEG-encoded string and returns a tensor suitable as input to the
// inception model.
func constructGraphToNormalizeImage(H, W int32, mean, scale float32, toBGR bool) (graph *tf.Graph, input, output tf.Output, err error) {
	// - The model was trained after with images scaled to scale*scale pixels.
	// - The colors, represented as R, G, B in 1-byte each were converted to
	//   float using (value - Mean)/Scale.

	// - input is a String-Tensor, where the string the JPEG-encoded image. PNG also supported
	// - The inception model takes a 4D tensor of shape
	//   [BatchSize, Height, Width, Colors=3], where each pixel is
	//   represented as a triplet of floats
	// - Apply normalization on each pixel and use ExpandDims to make
	//   this single image be a "batch" of size 1 for ResizeBilinear.

	// toBGR indicated whether changing RGB order to BGR
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s,
						op.DecodeJpeg(s, input, op.DecodeJpegChannels(3)), tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), mean)),
		op.Const(s.SubScope("scale"), scale))
	// RGB to BGR
	if toBGR {
		output = op.ReverseV2(s, output, op.Const(s, []int32{-1}))
	}
	graph, err = s.Finalize()
	return graph, input, output, err
}

// Convert the image bytes to a Tensor suitable as input
func makeTensorFromBytes(bytes []byte, H, W int32, mean, scale float32, toBGR bool) (*tf.Tensor, error) {
	// bytes to tensor
	tensor, err := tf.NewTensor(string(bytes))
	if err != nil {
		return nil, err
	}

	// create batch
	graph, input, output, err := constructGraphToNormalizeImage(H, W, mean, scale, toBGR)
	if err != nil {
		return nil, err
	}

	// Execute that graph create the batch of that image
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}

	defer session.Close()

	batch, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return batch[0], nil
}




// 计算box和旋转角度
func cropBox(imageByte []byte, box1 []float32) (*image.NRGBA, error) {
	var x1, y1, x2, y2 float32

	reader := bytes.NewReader(imageByte)

	img, err := imaging.Decode(reader)
	if err!=nil {
		return nil, err
	}

	w := float32(img.Bounds().Dx())
	h := float32(img.Bounds().Dy())

	box1[0] *= w
	box1[1] *= h
	box1[2] *= w
	box1[3] *= h

	// 计算需选择角度
	rotate_angle := 0

	if box1[0]<box1[2] { // 起点 在左
		if box1[1]<box1[3] { // 起点 在上
			rotate_angle = 0
			x1, y1, x2, y2 = box1[0], box1[1], box1[2], box1[3]
		} else {
			rotate_angle = 90
			x1, y1, x2, y2 = box1[0], box1[3], box1[2], box1[1]
		}
	} else{ // 起点 在右
		if box1[1]<box1[3] { // 起点 在上
			rotate_angle = 270
			x1, y1, x2, y2 = box1[2], box1[1], box1[0], box1[3]
		} else {
			rotate_angle = 180
			x1, y1, x2, y2 = box1[2], box1[3], box1[0], box1[1]
		}
	}

	if math.Abs(float64(x1-x2))<12 || math.Abs(float64(y1-y2))<12 { // 没有结果
		return nil, nil
	}

	return cropAndRotate(img, []int{int(x1), int(y1), int(x2), int(y2)}, rotate_angle), nil
}

// 挖出局部图片，并旋转
func cropAndRotate(src image.Image, box []int, rotate_angle int) *image.NRGBA {
	// 截取的框
	sr := image.Rectangle{
		image.Point{box[0], box[1]}, 
		image.Point{box[2], box[3]},
	}

	// 截取
	src2 := imaging.Crop(src, sr)

	//_ = imaging.Save(src2, "data/test1.jpg")

	// 旋转
	switch rotate_angle {
	case 90:
		src2 = imaging.Rotate270(src2)
	case 180:
		src2 = imaging.Rotate180(src2)
	case 270:
		src2 = imaging.Rotate90(src2)
	}

	//_ = imaging.Save(src2, "data/test2.jpg")

	return src2
}



// image 转换为 bytes
func image2bytes(img image.Image) ([]byte, error) {
	// 转换为 []byte
	buf := new(bytes.Buffer)
	err := imaging.Encode(buf, img, imaging.JPEG)
	if err!=nil {
		return nil, err
	}

	return buf.Bytes(), nil	
}


// 调整为方形，黑色填充
func padBox(src image.Image) *image.NRGBA {
	var maxW int

	if src.Bounds().Dx() > src.Bounds().Dy() {
		maxW = src.Bounds().Dx()
	} else {
		maxW = src.Bounds().Dy()
	}

	dst := imaging.New(maxW, maxW, color.Black)
	dst = imaging.PasteCenter(dst, src)

	//_ = imaging.Save(dst, "data/test3.jpg")

	return dst
}


// 概率转换为结果标签
func bestLabel(probabilities []float32) string{
	bestIdx := 0
	for i, p := range probabilities {
		if p > probabilities[bestIdx] {
			bestIdx = i
		}
	}

	return detposLabels[bestIdx]
}
