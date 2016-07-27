package mnistdemo

import (
	"encoding/json"
	"log"
	"math"
	"strconv"

	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/boosting"
)

const (
	stumpsStepCount    = 300
	stumpsCutoffCount  = 5
	stumpsSerializerID = "github.com/unixpickle/mnistdemo.Stumps"
)

func init() {
	serializer.RegisterTypedDeserializer(stumpsSerializerID, DeserializeStumps)
}

type Stump struct {
	Weight    float64
	Threshold float64
	X         int
	Y         int
}

func (s *Stump) Classify(b boosting.SampleList) linalg.Vector {
	res := make(linalg.Vector, b.Len())
	for i, sample := range b.(stumpSampleList) {
		if s.ClassifySingle(sample.Sample) {
			res[i] = s.Weight
		} else {
			res[i] = -s.Weight
		}
	}
	return res
}

func (s *Stump) ClassifySingle(sample *Sample) bool {
	return sample[s.X+s.Y*28] > s.Threshold
}

type Stumps struct {
	Stumps map[string][]*Stump
}

func DeserializeStumps(d []byte) (*Stumps, error) {
	dec, err := decompress(d)
	if err != nil {
		return nil, err
	}
	var res Stumps
	if err := json.Unmarshal(dec, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Stumps) Train(data, validation []*TrainingSample) {
	s.Stumps = map[string][]*Stump{}

	log.Println("Creating stump pool...")
	pool := createStumpPool(stumpSampleList(data))
	for digit := 0; digit < 10; digit++ {
		log.Printf("Learning stumps for %d...", digit)
		classVec := make(linalg.Vector, len(data))
		for i, x := range data {
			if x.Label == digit {
				classVec[i] = 1
			} else {
				classVec[i] = -1
			}
		}
		grad := boosting.Gradient{
			Loss:    boosting.ExpLoss{},
			Desired: classVec,
			List:    stumpSampleList(data),
			Pool:    pool,
		}
		for i := 0; i < stumpsStepCount; i++ {
			grad.Step()
		}

		var stumpList []*Stump
		for i, stump := range grad.Sum.Classifiers {
			sCopy := *stump.(*Stump)
			sCopy.Weight = grad.Sum.Weights[i]
			stumpList = append(stumpList, &sCopy)
		}
		s.Stumps[strconv.Itoa(digit)] = stumpList
	}

	log.Println("Validating...")
	var correct int
	for _, sample := range validation {
		if s.Classify(sample.Sample) == sample.Label {
			correct++
		}
	}
	log.Printf("Validation results: %d/%d", correct, len(validation))
}

func (s *Stumps) Classify(sample *Sample) int {
	bestSum := math.Inf(-1)
	var bestDigit string
	for digit, stumps := range s.Stumps {
		var sum float64
		for _, stump := range stumps {
			if stump.ClassifySingle(sample) {
				sum += stump.Weight
			} else {
				sum -= stump.Weight
			}
		}
		if sum > bestSum {
			bestSum = sum
			bestDigit = digit
		}
	}
	res, _ := strconv.Atoi(bestDigit)
	return res
}

func (s *Stumps) SerializerType() string {
	return stumpsSerializerID
}

func (s *Stumps) Serialize() ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return compress(data), nil
}

func createStumpPool(s boosting.SampleList) boosting.Pool {
	var classifiers []boosting.Classifier
	divide := 1 / float64(stumpsCutoffCount+1)
	for y := 0; y < 28; y++ {
		for x := 0; x < 28; x++ {
			for cutoff := 0; cutoff < stumpsCutoffCount; cutoff++ {
				thresh := divide * float64(cutoff+1)
				classifiers = append(classifiers, &Stump{
					Weight:    1,
					Threshold: thresh,
					X:         x,
					Y:         y,
				})
			}
		}
	}
	return boosting.NewStaticPool(classifiers, s)
}

type stumpSampleList []*TrainingSample

func (s stumpSampleList) Len() int {
	return len(s)
}
