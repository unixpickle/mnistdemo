package mnistdemo

import (
	"encoding/json"
	"log"
	"math"

	"github.com/unixpickle/serializer"
)

const bayesSerializerID = "github.com/unixpickle/mnistdemo.Bayes"

func init() {
	serializer.RegisterTypedDeserializer(bayesSerializerID, DeserializeBayes)
}

type Gaussian struct {
	Mean     float64
	Variance float64
}

type Bayes struct {
	Classes [10][28 * 28]Gaussian
	Total   [28 * 28]Gaussian
}

func DeserializeBayes(d []byte) (*Bayes, error) {
	var bayes Bayes
	if err := json.Unmarshal(d, &bayes); err != nil {
		return nil, err
	}
	return &bayes, nil
}

func (b *Bayes) Train(data, validation []*TrainingSample) {
	log.Println("Training classifier...")
	for i := 0; i < 10; i++ {
		computeGaussians(&b.Classes[i], data, func(j int) bool {
			return j == i
		})
	}
	computeGaussians(&b.Total, data, func(j int) bool {
		return true
	})
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
	for i := 0; i < 10; i++ {
		gaussians := &b.Classes[i]
		var logProb float64
		for j, g := range gaussians[:] {
			g0 := b.Total[j]
			logProb += math.Log(g0.Variance) - math.Log(g.Variance)
			logProb += math.Pow(s[j]-g0.Mean, 2) / g0.Variance
			logProb -= math.Pow(s[j]-g.Mean, 2) / g.Variance
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
	return json.Marshal(b)
}

func computeGaussians(g *[28 * 28]Gaussian, set []*TrainingSample, filter func(n int) bool) {
	var total int
	for _, x := range set {
		if filter(x.Label) {
			total++
			for i, v := range x.Sample {
				g[i].Mean += v
			}
		}
	}
	for i := range g[:] {
		g[i].Mean /= float64(total)
	}
	for _, x := range set {
		if filter(x.Label) {
			for i, v := range x.Sample {
				diff := v - g[i].Mean
				g[i].Variance += diff * diff
			}
		}
	}
	for i := range g[:] {
		g[i].Variance /= float64(total)
		if g[i].Variance == 0 {
			g[i].Variance = 1.0 / 100.0
		}
	}
}
