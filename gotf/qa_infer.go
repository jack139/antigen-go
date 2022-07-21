package gotf

import (
	//"log"
	"strings"

	"github.com/buckhx/gobert/tokenize"
	"github.com/buckhx/gobert/tokenize/vocab"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/aclements/go-gg/generic/slice"
)

const (
	/* 预训练模型路径 */
	modelPath = "../../nlp/qa_demo/saved-model"
	vocabPath = "../../nlp/qa_demo/saved-model/vocab_chinese.txt"

	MaxSeqLength = 512
)

/* 训练好的模型权重 */
var m *tf.SavedModel
var voc vocab.Dict


/* 初始化模型 */
func InitModel() error {
	var err error
	voc, err = vocab.FromFile(vocabPath)
	if err != nil {
		return err
	}
	m, err = tf.LoadSavedModel(modelPath, []string{"train"}, nil)
	if err != nil {
		return err
	}

	return nil
}

/* 判断是否是英文字符 */
func isAlpha(c byte) bool {
	return (c>=65 && c<=90) || (c>=97 && c<=122)
}


func BertQA(corpus string, question string) (ans string, err error) {

	//log.Printf("Corpus: %s\tQuestion: %s", corpus, question)

	tkz := tokenize.NewTokenizer(voc)
	ff := tokenize.FeatureFactory{Tokenizer: tkz, SeqLen: MaxSeqLength}
	// 拼接输入
	input_tokens := question + tokenize.SequenceSeparator + corpus
	// 获取 token 向量
	f := ff.Feature(input_tokens)

	tids, err := tf.NewTensor([][]int32{f.TokenIDs})
	if err != nil {
		return ans, err
	}
	new_mask := make([]float32, len(f.Mask))
	for i, v := range f.Mask {
		new_mask[i] = float32(v)
	}
	mask, err := tf.NewTensor([][]float32{new_mask})
	if err != nil {
		return ans, err
	}
	sids, err := tf.NewTensor([][]int32{f.TypeIDs})
	if err != nil {
		return ans, err
	}

	res, err := m.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.Graph.Operation("input_ids").Output(0):      tids,
			m.Graph.Operation("input_mask").Output(0):     mask,
			m.Graph.Operation("segment_ids").Output(0):    sids,
		},
		[]tf.Output{
			m.Graph.Operation("finetune_mrc/Squeeze").Output(0),
			m.Graph.Operation("finetune_mrc/Squeeze_1").Output(0),
		},
		nil,
	)
	if err != nil {
		return ans, err
	}

	st := slice.ArgMax(res[0].Value().([][]float32)[0])
	ed := slice.ArgMax(res[1].Value().([][]float32)[0])
	//fmt.Println(st, ed)
	if ed<st{ // ed 小于 st 说明未找到答案
		st = 0
		ed = 0
	}
	//ans = strings.Join(f.Tokens[st:ed+1], "")

	// 处理token中的英文，例如： 'di', '##st', '##ri', '##bu', '##ted', 're', '##pr', '##ese', '##nt', '##ation',
	ans = ""
	for i:=st;i<ed+1;i++ {
		if len(f.Tokens[i])>0 && isAlpha(f.Tokens[i][0]){ // 英文开头，加空格
			ans += " "+f.Tokens[i]
		} else if strings.HasPrefix(f.Tokens[i], "##"){ // ##开头，是英文中段，去掉##
			ans += f.Tokens[i][2:]
		} else {
			ans += f.Tokens[i]
		}
	}

	if strings.HasPrefix(ans, "[CLS]") || strings.HasPrefix(ans, "[SEP]") {
		return "", nil
	} else {
		return ans, nil // 找到答案
	}
}
