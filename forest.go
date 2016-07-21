package mnistdemo

import (
	"encoding/json"
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
	forestTreeCount    = 100
	forestSampleSubset = 4000
	forestAttrSubset   = 75
)

func init() {
	serializer.RegisterTypedDeserializer(forestSerializerID, DeserializeForest)
}

// A Forest is a random forest.
type Forest struct {
	F idtrees.Forest
}

// DeserializeForest deserializes a Forest that was
// previously serialized with Forest.Serialize().
func DeserializeForest(d []byte) (*Forest, error) {
	var archived []*archivedTree
	if err := json.Unmarshal(d, &archived); err != nil {
		return nil, err
	}
	res := &Forest{F: make(idtrees.Forest, len(archived))}
	for i, a := range archived {
		res.F[i] = unarchiveTree(a)
	}
	return res, nil
}

// Train trains the forest on the given training data.
func (f *Forest) Train(data, validation []*TrainingSample) {
	rand.Seed(time.Now().UnixNano())
	samples := newForestSamples(data)
	attrs := forestAttrs()
	log.Println("Building forest...")
	f.F = idtrees.BuildForest(forestTreeCount, samples, attrs, forestSampleSubset,
		forestAttrSubset,
		func(s []idtrees.Sample, a []idtrees.Attr) *idtrees.Tree {
			return idtrees.ID3(s, a, 0)
		})
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
	sample := forestSample{T: &TrainingSample{Sample: s}}
	res := f.F.Classify(sample)

	var maxClass int
	maxLikelihood := math.Inf(-1)
	for class, likelihood := range res {
		if likelihood > maxLikelihood {
			maxLikelihood = likelihood
			maxClass = class.(int)
		}
	}
	return maxClass
}

// SerializerType returns Forest's unique type ID
// to be used in the serializer database.
func (f *Forest) SerializerType() string {
	return forestSerializerID
}

// Serialize serializes the forest's data.
func (f *Forest) Serialize() ([]byte, error) {
	a := make([]*archivedTree, len(f.F))
	for i, t := range f.F {
		a[i] = archiveTree(t)
	}
	return json.Marshal(a)
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

func unarchiveTree(t *archivedTree) *idtrees.Tree {
	res := new(idtrees.Tree)
	if t.Classification != nil {
		res.Classification = map[idtrees.Class]float64{}
		for k, v := range t.Classification {
			class, _ := strconv.Atoi(k)
			res.Classification[class] = v
		}
		return res
	}
	res.Attr = t.Pixel
	res.NumSplit = &idtrees.NumSplit{
		Threshold: t.Threshold,
		LessEqual: unarchiveTree(t.LessEqual),
		Greater:   unarchiveTree(t.Greater),
	}
	return res
}
