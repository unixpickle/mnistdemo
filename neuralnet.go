package mnistdemo

import (
	"errors"
	"log"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
)

const (
	neuralnetSerializerID = "github.com/unixpickle/mnistdemo.NeuralNet"
	neuralnetFilterCount  = 8
)

func init() {
	serializer.RegisterTypedDeserializer(neuralnetSerializerID, DeserializeNeuralNet)
}

type NeuralNet struct {
	Net neuralnet.Network
}

func DeserializeNeuralNet(d []byte) (*NeuralNet, error) {
	dec, err := decompress(d)
	if err != nil {
		return nil, errors.New("failed to decompress network: " + err.Error())
	}
	nn, err := neuralnet.DeserializeNetwork(dec)
	if err != nil {
		return nil, err
	}
	return &NeuralNet{Net: nn}, nil
}

func NewNeuralNet() *NeuralNet {
	convLayer := &neuralnet.ConvLayer{
		FilterCount:  neuralnetFilterCount,
		FilterWidth:  3,
		FilterHeight: 3,
		Stride:       1,
		InputWidth:   28,
		InputHeight:  28,
		InputDepth:   1,
	}
	pool := &neuralnet.MaxPoolingLayer{
		XSpan:       3,
		YSpan:       3,
		InputWidth:  convLayer.OutputWidth(),
		InputHeight: convLayer.OutputHeight(),
		InputDepth:  neuralnetFilterCount,
	}
	net := neuralnet.Network{
		convLayer,
		&neuralnet.Sigmoid{},
		pool,
		&neuralnet.DenseLayer{
			InputCount: pool.OutputWidth() * pool.OutputHeight() *
				neuralnetFilterCount,
			OutputCount: 300,
		},
		&neuralnet.Sigmoid{},
		&neuralnet.DenseLayer{
			InputCount:  300,
			OutputCount: 10,
		},
		&neuralnet.LogSoftmaxLayer{},
	}
	net.Randomize()
	return &NeuralNet{Net: net}
}

func (n *NeuralNet) Train(data, validation []*TrainingSample) {
	log.Println("Training classifier (ctrl+c to stop)...")

	samples := neuralnetSampleSet(data)
	gradienter := &neuralnet.BatchRGradienter{
		Learner:  n.Net.BatchLearner(),
		CostFunc: neuralnet.DotCost{},
	}
	adam := &sgd.Adam{Gradienter: gradienter}
	sgd.SGDInteractive(adam, samples, 0.001, 50, func() bool {
		log.Println("Mid-training score:", n.score(validation))
		return true
	})

	log.Println("Running cross validation...")
	var correct, total int
	for _, s := range validation {
		total++
		if n.Classify(s.Sample) == s.Label {
			correct++
		}
	}
	log.Printf("Got %d/%d", correct, total)
}

func (n *NeuralNet) Classify(s *Sample) int {
	inVar := &autofunc.Variable{Vector: s[:]}
	output := n.Net.Apply(inVar).Output()

	var greatestVal float64
	var greatestIdx int
	for i, x := range output {
		if i == 0 || x > greatestVal {
			greatestVal = x
			greatestIdx = i
		}
	}

	return greatestIdx
}

func (n *NeuralNet) SerializerType() string {
	return neuralnetSerializerID
}

func (n *NeuralNet) Serialize() ([]byte, error) {
	raw, err := n.Net.Serialize()
	if err != nil {
		return nil, err
	}
	return compress(raw), nil
}

func (n *NeuralNet) score(v []*TrainingSample) float64 {
	var correct, total int
	for _, s := range v {
		total++
		if n.Classify(s.Sample) == s.Label {
			correct++
		}
	}
	return float64(correct) / float64(total)
}

func neuralnetSampleSet(data []*TrainingSample) sgd.SampleSet {
	var inputVecs, labelVecs []linalg.Vector
	for _, t := range data {
		inVec, labelVec := neuralnetSample(t)
		inputVecs = append(inputVecs, inVec)
		labelVecs = append(labelVecs, labelVec)
	}
	return neuralnet.VectorSampleSet(inputVecs, labelVecs)
}

func neuralnetSample(t *TrainingSample) (inVec, outVec linalg.Vector) {
	inVec = t.Sample[:]
	outVec = make(linalg.Vector, 10)
	outVec[t.Label] = 1
	return
}
