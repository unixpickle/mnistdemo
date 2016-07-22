package mnistdemo

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/unixpickle/serializer"
	"github.com/unixpickle/weakai/idtrees"
)

const (
	forestSerializerID = "github.com/unixpickle/mnistdemo.Forest"
	forestTreeCount    = 70
	forestSampleSubset = 4000
	forestAttrSubset   = 75
)

func init() {
	serializer.RegisterTypedDeserializer(forestSerializerID, DeserializeForest)
}

// A Forest is a random forest.
type Forest struct {
	F []*archivedTree
}

// DeserializeForest deserializes a Forest that was
// previously serialized with Forest.Serialize().
func DeserializeForest(compressed []byte) (*Forest, error) {
	d, err := decompress(compressed)
	if err != nil {
		return nil, errors.New("failed to decompress tree: " + err.Error())
	}
	var archived []*archivedTree
	if err := unmarshalTrees(d, &archived); err != nil {
		return nil, err
	}
	return &Forest{F: archived}, nil
}

// Train trains the forest on the given training data.
func (f *Forest) Train(data, validation []*TrainingSample) {
	rand.Seed(time.Now().UnixNano())
	samples := newForestSamples(data)
	attrs := forestAttrs()
	log.Println("Building forest...")
	forest := idtrees.BuildForest(forestTreeCount, samples, attrs, forestSampleSubset,
		forestAttrSubset,
		func(s []idtrees.Sample, a []idtrees.Attr) *idtrees.Tree {
			return idtrees.ID3(s, a, 0)
		})
	f.F = make([]*archivedTree, len(forest))
	for i, t := range forest {
		f.F[i] = archiveTree(t)
	}
	log.Println("Running cross validation...")
	var correctCount int
	for _, vSamp := range validation {
		actual := f.Classify(vSamp.Sample)
		if actual == vSamp.Label {
			correctCount++
		}
	}
	log.Printf("Validation score: %d/%d", correctCount, len(validation))
}

// Classify returns the most likely class for the sample.
func (f *Forest) Classify(s *Sample) int {
	sums := map[string]float64{}
	for _, t := range f.F {
		res := t.Classify(s)
		for key, val := range res {
			sums[key] += val
		}
	}

	maxClass := "0"
	maxLikelihood := math.Inf(-1)
	for class, likelihood := range sums {
		if likelihood > maxLikelihood {
			maxLikelihood = likelihood
			maxClass = class
		}
	}

	res, _ := strconv.Atoi(maxClass)
	return res
}

// SerializerType returns Forest's unique type ID
// to be used in the serializer database.
func (f *Forest) SerializerType() string {
	return forestSerializerID
}

// Serialize serializes the forest's data.
func (f *Forest) Serialize() ([]byte, error) {
	data, err := json.Marshal(f.F)
	if err != nil {
		return nil, err
	}
	return compress(data), nil
}

func forestAttrs() []idtrees.Attr {
	res := make([]idtrees.Attr, 28*28)
	for i := range res {
		res[i] = i
	}
	return res
}

type forestSample struct {
	T *TrainingSample
}

func newForestSamples(ts []*TrainingSample) []idtrees.Sample {
	res := make([]idtrees.Sample, len(ts))
	for i, t := range ts {
		res[i] = forestSample{T: t}
	}
	return res
}

func (f forestSample) Attr(a idtrees.Attr) idtrees.Val {
	return f.T.Sample[a.(int)]
}

func (f forestSample) Class() idtrees.Class {
	return f.T.Label
}

type archivedTree struct {
	Classification map[string]float64
	Pixel          int
	Threshold      float64
	LessEqual      *archivedTree
	Greater        *archivedTree
}

func archiveTree(t *idtrees.Tree) *archivedTree {
	res := new(archivedTree)
	if t.Classification != nil {
		res.Classification = map[string]float64{}
		for k, v := range t.Classification {
			res.Classification[strconv.Itoa(k.(int))] = v
		}
		return res
	}
	res.Pixel = t.Attr.(int)
	res.Threshold = t.NumSplit.Threshold.(float64)
	res.LessEqual = archiveTree(t.NumSplit.LessEqual)
	res.Greater = archiveTree(t.NumSplit.Greater)
	return res
}

func (a *archivedTree) Classify(s *Sample) map[string]float64 {
	for a.Classification == nil {
		p := s[a.Pixel]
		if p > a.Threshold {
			a = a.Greater
		} else {
			a = a.LessEqual
		}
	}
	return a.Classification
}
