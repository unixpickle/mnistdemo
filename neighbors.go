package mnistdemo

import (
	"bytes"
	"encoding/gob"
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/unixpickle/serializer"
)

const (
	neighborSampleCount = 500
	neighborsMaxK       = 30

	neighborsSerializerID = "github.com/unixpickle/mnistdemo.Neighbors"
)

func init() {
	serializer.RegisterTypedDeserializer(neighborsSerializerID, DeserializeNeighbors)
}

type Neighbors struct {
	Images [10][][]byte
	K      int
}

func DeserializeNeighbors(d []byte) (*Neighbors, error) {
	dec, err := decompress(d)
	if err != nil {
		return nil, err
	}
	gobReader := gob.NewDecoder(bytes.NewBuffer(dec))
	var res Neighbors
	if err := gobReader.Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (n *Neighbors) Train(data, validation []*TrainingSample) {
	log.Println("Choosing samples...")
	for i := 0; i < 10; i++ {
		n.Images[i] = neighborSamples(data, i)
	}
	log.Println("Selecting K value...")
	kScores := map[int]int{}
	for _, sample := range validation {
		res := n.resultsForSample(sample.Sample)
		m := map[int]int{}
		for k := 1; k <= neighborsMaxK; k++ {
			m[res.Labels[k-1]]++
			if keyForMaxCount(m) == sample.Label {
				kScores[k]++
			}
		}
	}
	n.K = keyForMaxCount(kScores)
	log.Printf("For k=%d score is %d/%d...", n.K, kScores[n.K], len(validation))
}

func (n *Neighbors) Classify(s *Sample) int {
	res := n.resultsForSample(s)
	counts := map[int]int{}
	for i := 0; i < n.K; i++ {
		counts[res.Labels[i]]++
	}
	return keyForMaxCount(counts)
}

func (n *Neighbors) SerializerType() string {
	return neighborsSerializerID
}

func (n *Neighbors) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(n); err != nil {
		return nil, err
	}
	return compress(buf.Bytes()), nil
}

func (n *Neighbors) crossScore(validation []*TrainingSample) int {
	var correctCount int
	for _, sample := range validation {
		if n.Classify(sample.Sample) == sample.Label {
			correctCount++
		}
	}
	return correctCount
}

func (n *Neighbors) resultsForSample(s *Sample) *classifierResults {
	var res classifierResults
	for label, examples := range n.Images[:] {
		for _, example := range examples {
			dist := 1 - cosineDistance(s, example)
			res.Distances = append(res.Distances, dist)
			res.Labels = append(res.Labels, label)
		}
	}
	sort.Sort(&res)
	return &res
}

func cosineDistance(s *Sample, template []byte) float64 {
	var dotProduct float64
	var sampleMag float64
	var templateMag float64
	for i, x := range s {
		y := float64(template[i]) / 255
		dotProduct += x * y
		sampleMag += x * x
		templateMag += y * y
	}
	return dotProduct / math.Sqrt(sampleMag*templateMag)
}

func neighborSamples(data []*TrainingSample, label int) [][]byte {
	var allSamples [][]byte
	for _, x := range data {
		if x.Label == label {
			byteList := make([]byte, len(x.Sample))
			for i, f := range x.Sample[:] {
				byteList[i] = byte(f*255 + 0.5)
			}
			allSamples = append(allSamples, byteList)
		}
	}
	perm := rand.Perm(len(allSamples))

	res := make([][]byte, neighborSampleCount)
	for i, j := range perm[:len(res)] {
		res[i] = allSamples[j]
	}
	return res
}

func keyForMaxCount(m map[int]int) int {
	var bestKey int
	bestCount := -1
	for key, val := range m {
		if val > bestCount {
			bestCount = val
			bestKey = key
		}
	}
	return bestKey
}

type classifierResults struct {
	Labels    []int
	Distances []float64
}

func (c *classifierResults) Len() int {
	return len(c.Labels)
}

func (c *classifierResults) Swap(i, j int) {
	c.Labels[i], c.Labels[j] = c.Labels[j], c.Labels[i]
	c.Distances[i], c.Distances[j] = c.Distances[j], c.Distances[i]
}

func (c *classifierResults) Less(i, j int) bool {
	return c.Distances[i] < c.Distances[j]
}
