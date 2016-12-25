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
	bayesEigIterations = 300
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
	Total   [bayesFeatures]Gaussian
	Basis   *linalg.Matrix
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
	log.Println("Computing basis features...")
	b.computeBasis(data)
	log.Println("Training classifier...")
	for i := 0; i < 10; i++ {
		b.computeGaussians(&b.Classes[i], data, func(j int) bool {
			return j == i
		})
	}
	b.computeGaussians(&b.Total, data, func(j int) bool {
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
	features := b.Basis.MulFast(linalg.NewMatrixColumn(s[:])).Data
	bestLogProb := math.Inf(-1)
	bestAnswer := 0
	for i := 0; i < 10; i++ {
		gaussians := &b.Classes[i]
		var logProb float64
		for j, g := range gaussians[:] {
			g0 := b.Total[j]
			logProb += math.Log(g0.Variance) - math.Log(g.Variance)
			logProb += math.Pow(features[j]-g0.Mean, 2) / g0.Variance
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

func (b *Bayes) computeBasis(data []*TrainingSample) {
	log.Println("Computing covariance matrix...")
	cvm := computeCovarianceMatrix(data)
	log.Println("Computing eigenvalues...")
	vals, vecs := largestEigenvectors(cvm)
	log.Println("Largest variance:", linalg.Vector(vals).MaxAbs())

	basis := vecs[:bayesFeatures]
	b.Basis = linalg.NewMatrix(bayesFeatures, 28*28)
	for i, x := range basis {
		copy(b.Basis.Data[i*28*28:(i+1)*28*28], x)
	}
}

func (b *Bayes) computeGaussians(g *[bayesFeatures]Gaussian, set []*TrainingSample,
	filter func(n int) bool) {
	var total int
	for _, x := range set {
		if filter(x.Label) {
			total++
			features := b.Basis.MulFast(linalg.NewMatrixColumn(x.Sample[:]))
			for i, v := range features.Data {
				g[i].Mean += v
				g[i].Variance += v * v
			}
		}
	}
	for i := range g[:] {
		g[i].Mean /= float64(total)
		g[i].Variance = g[i].Variance/float64(total) - g[i].Mean*g[i].Mean
	}
}

func computeCovarianceMatrix(data []*TrainingSample) *linalg.Matrix {
	return approb.Covariances(5000, func() linalg.Vector {
		return data[rand.Intn(len(data))].Sample[:]
	})
}

func largestEigenvectors(mat *linalg.Matrix) (vals []float64, vecs []linalg.Vector) {
	vecMat := linalg.NewMatrix(28*28, bayesFeatures)
	for i := range vecMat.Data {
		vecMat.Data[i] = rand.NormFloat64()
	}
	for i := 0; i < bayesEigIterations; i++ {
		product := mat.MulFast(vecMat)
		vecMat, _ = qrdecomp.Householder(product)
	}
	finalProduct := mat.MulFast(vecMat)
	for i := 0; i < bayesFeatures; i++ {
		col := finalProduct.Col(i)
		vals = append(vals, col.Mag())
		vecs = append(vecs, col)
	}
	return
}
