package detpos

import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	"log"
	"os"
	"image/draw"
	"bytes"
	"image"
	"image/jpeg"
	"math"
)

/*
	below codes taken from
	https://github.com/tensorflow/tensorflow/blob/master/tensorflow/go/example_inception_inference_test.go
*/

// This function constructs a graph of TensorFlow operations which takes as
// input a JPEG-encoded string and returns a tensor suitable as input to the
// inception model.
func constructGraphToNormalizeImage() (graph *tf.Graph, input, output tf.Output, err error) {
	// Some constants specific to the pre-trained model at:
	// https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip
	//
	// - The model was trained after with images scaled to 224x224 pixels.
	// - The colors, represented as R, G, B in 1-byte each were converted to
	//   float using (value - Mean)/Scale.
	const (
		H, W  = 256, 256
		Mean  = float32(0)
		Scale = float32(255)
	)
	// - input is a String-Tensor, where the string the JPEG-encoded image.
	// - The inception model takes a 4D tensor of shape
	//   [BatchSize, Height, Width, Colors=3], where each pixel is
	//   represented as a triplet of floats
	// - Apply normalization on each pixel and use ExpandDims to make
	//   this single image be a "batch" of size 1 for ResizeBilinear.
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.ReverseV2(s, 
		op.Div(s,
			op.Sub(s,
				op.ResizeBilinear(s,
					op.ExpandDims(s,
						op.Cast(s,
							op.DecodeJpeg(s, input, op.DecodeJpegChannels(3)), tf.Float),
						op.Const(s.SubScope("make_batch"), int32(0))),
					op.Const(s.SubScope("size"), []int32{H, W})),
				op.Const(s.SubScope("mean"), Mean)),
			op.Const(s.SubScope("scale"), Scale)), 
		op.Const(s, []int32{-1}))
	graph, err = s.Finalize()
	return graph, input, output, err
}

// Convert the image in filename to a Tensor suitable as input
func makeTensorFromBytes(bytes []byte) (*tf.Tensor, error) {
	// bytes to tensor
	tensor, err := tf.NewTensor(string(bytes))
	if err != nil {
		return nil, err
	}

	// create batch
	graph, input, output, err := constructGraphToNormalizeImage()
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
func cropBox(imageByte []byte, box1 []float32) ([]int, int) {
	var x1, y1, x2, y2 float32

	reader := bytes.NewReader(imageByte)

	img, _, _ := image.Decode(reader)

	log.Printf("%v", img.Bounds())

	w := float32(img.Bounds().Dx())
	h := float32(img.Bounds().Dy())

	box1[0] *= w
	box1[1] *= h
	box1[2] *= w
	box1[3] *= h

	log.Println("box1: ", box1)

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

	//x1, y1, x2, y2 = int(x1), int(y1), int(x2), int(y2)

	if math.Abs(float64(x1-x2))<12 || math.Abs(float64(y1-y2))<12 { // 没有结果
		return nil, 0
	}

	crop := cropImage(&img, []int{int(x1), int(y1), int(x2), int(y2)}, rotate_angle)

	saveImage("data/test.jpg", crop)

	return []int{int(x1), int(y1), int(x2), int(y2)}, rotate_angle
}

// 挖出局部图片，并旋转
func cropImage(src *image.Image, box []int, rotate_angle int) *image.RGBA {

	log.Println("crop box: ", box)

	dst := image.NewRGBA(image.Rect(0, 0, box[2]-box[0], box[3]-box[1]))

	dp := dst.Bounds().Min
	sr := image.Rectangle{
		image.Point{box[0], box[1]}, 
		image.Point{box[2], box[3]},
	}
	r := image.Rectangle{dp, dp.Add(sr.Size())}
	draw.Draw(dst, r, *src, sr.Min, draw.Src)

	return dst
}

func saveImage(filename string, img *image.RGBA){
	toimg, _ := os.Create(filename)
	defer toimg.Close()

	jpeg.Encode(toimg, img, &jpeg.Options{jpeg.DefaultQuality})
}