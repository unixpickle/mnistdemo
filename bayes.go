package mnistdemo

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"

	"github.com/unixpickle/approb"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/num-analysis/linalg/qrdecomp"
	"github.com/unixpickle/serializer"
)

const bayesSerializerID = "github.com/unixpickle/mnistdemo.Bayes"

const (
	bayesFeatures      = 50
	bayesEigIterations = 200
	bayesCovSamples    = 2000
)

func init() {
	serializer.RegisterTypedDeserializer(bayesSerializerID, DeserializeBayes)
}

type Gaussian struct {
	Mean     float64
	Variance float64
}

type Bayes struct {
	Classes [10][bayesFeatures]Gaussian
	Bases   [10]*linalg.Matrix
}

func DeserializeBayes(d []byte) (*Bayes, error) {
	var bayes Bayes
	data, err := decompress(d)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &bayes); err != nil {
		return nil, err
	}
	return &bayes, nil
}

func (b *Bayes) Train(data, validation []*TrainingSample) {
	log.Println("Training classifier...")
	for i := 0; i < 10; i++ {
		log.Println("Label", i)
		b.computeBasis(data, i)
		b.computeGaussians(data, i)
	}
	log.Println("Running cross validation...")
	var correct, total int
	for _, s := range validation {
		total++
		if b.Classify(s.Sample) == s.Label {
			correct++
		}
	}
	log.Printf("Got %d/%d", correct, total)
}

func (b *Bayes) Classify(s *Sample) int {
	bestLogProb := math.Inf(-1)
	bestAnswer := 0
	for i, basis := range b.Bases[:] {
		features := basis.MulFast(linalg.NewMatrixColumn(s[:])).Data
		gaussians := &b.Classes[i]
		var logProb float64
		for j, g := range gaussians[:] {
			logProb -= math.Log(g.Variance)
			logProb -= math.Pow(features[j]-g.Mean, 2) / g.Variance
		}
		if logProb > bestLogProb {
			bestLogProb = logProb
			bestAnswer = i
		}
	}
	return bestAnswer
}

func (b *Bayes) SerializerType() string {
	return bayesSerializerID
}

func (b *Bayes) Serialize() ([]byte, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return compress(data), nil
}

func (b *Bayes) computeBasis(data []*TrainingSample, label int) {
	cvm := computeCovarianceMatrix(data, label)
	b.Bases[label] = largestEigenvectors(cvm)
}

func (b *Bayes) computeGaussians(set []*TrainingSample, label int) {
	g := &b.Classes[label]
	var total int
	for _, x := range set {
		if x.Label != label {
			continue
		}
		total++
		features := b.Bases[label].MulFast(linalg.NewMatrixColumn(x.Sample[:]))
		for i, v := range features.Data {
			g[i].Mean += v
			g[i].Variance += v * v
		}
	}
	for i := range g[:] {
		g[i].Mean /= float64(total)
		g[i].Variance = g[i].Variance/float64(total) - g[i].Mean*g[i].Mean
	}
}

func computeCovarianceMatrix(data []*TrainingSample, label int) *linalg.Matrix {
	return approb.Covariances(bayesCovSamples, func() linalg.Vector {
		for {
			item := data[rand.Intn(len(data))]
			if item.Label == label {
				return item.Sample[:]
			}
		}
	})
}

func largestEigenvectors(mat *linalg.Matrix) *linalg.Matrix {
	vecMat := linalg.NewMatrix(28*28, bayesFeatures)
	for i := range vecMat.Data {
		vecMat.Data[i] = rand.NormFloat64()
	}
	for i := 0; i < bayesEigIterations; i++ {
		product := mat.MulFast(vecMat)
		vecMat, _ = qrdecomp.Householder(product)
	}
	return vecMat.Transpose()
}
