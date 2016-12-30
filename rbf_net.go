package mnistdemo

import (
	"errors"
	"log"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rbf"
)

const (
	rbfNetSerializerID = "github.com/unixpickle/mnistdemo.RBFNet"
	rbfNetFilterCount  = 8
)

func init() {
	serializer.RegisterTypedDeserializer(rbfNetSerializerID, DeserializeRBFNet)
}

type RBFNet struct {
	Net *rbf.Network
}

func DeserializeRBFNet(d []byte) (*RBFNet, error) {
	dec, err := decompress(d)
	if err != nil {
		return nil, errors.New("failed to decompress network: " + err.Error())
	}
	n, err := rbf.DeserializeNetwork(dec)
	if err != nil {
		return nil, err
	}
	return &RBFNet{Net: n}, nil
}

func (n *RBFNet) Train(data, validation []*TrainingSample) {
	log.Println("Initializing network...")
	samples := neuralnetSampleSet(data)
	n.Net = &rbf.Network{
		DistLayer:  rbf.NewDistLayerSamples(28*28, 300, samples),
		ScaleLayer: rbf.NewScaleLayerShared(0.05),
		ExpLayer:   &rbf.ExpLayer{Normalize: true},
	}

	log.Println("Least-squares pre-training...")
	sgd.ShuffleSampleSet(samples)
	n.Net.OutLayer = rbf.LeastSquares(n.Net, samples.Subset(0, 10000), 20)

	log.Println("Fine-tuning with SGD...")

	gradienter := &neuralnet.BatchRGradienter{
		Learner:  n.Net,
		CostFunc: neuralnet.MeanSquaredCost{},
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

func (n *RBFNet) Classify(s *Sample) int {
	inVar := &autofunc.Variable{Vector: s[:]}
	output := n.Net.Apply(inVar).Output()
	_, idx := output.Max()
	return idx
}

func (n *RBFNet) SerializerType() string {
	return rbfNetSerializerID
}

func (n *RBFNet) Serialize() ([]byte, error) {
	raw, err := n.Net.Serialize()
	if err != nil {
		return nil, err
	}
	return compress(raw), nil
}

func (n *RBFNet) score(v []*TrainingSample) float64 {
	var correct, total int
	for _, s := range v {
		total++
		if n.Classify(s.Sample) == s.Label {
			correct++
		}
	}
	return float64(correct) / float64(total)
}
