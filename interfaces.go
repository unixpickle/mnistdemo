package mnistdemo

import "github.com/unixpickle/serializer"

// Sample is a 28x28 image.
// Pixels in the image are 1 if they're black and
// 0 if they're white.
type Sample [28 * 28]float64

// A TrainingSample is an image sample with a
// corresponding digit label.
type TrainingSample struct {
	Sample *Sample
	Label  int
}

// A Classifier can be trained on a set of samples
// and predicts labels for samples.
type Classifier interface {
	serializer.Serializer

	Train(data, validation []*TrainingSample)
	Classify(s *Sample) int
}

// A ClassifierDesc includes a plain-text description
// of a classifier as well as a constructor for that
// classifier.
type ClassifierDesc struct {
	Desc      string
	Construct func() Classifier
}

// Classifiers stores ClassifierDescs for each available
// classifier.
var Classifiers = map[string]ClassifierDesc{
	"forest": ClassifierDesc{
		Desc: "random forests with ID3",
		Construct: func() Classifier {
			return &Forest{}
		},
	},
	"bayes": ClassifierDesc{
		Desc: "naive bayes classification",
		Construct: func() Classifier {
			return &Bayes{}
		},
	},
	"neuralnet": ClassifierDesc{
		Desc: "a basic convolutional net",
		Construct: func() Classifier {
			return NewNeuralNet()
		},
	},
	"neighbors": ClassifierDesc{
		Desc: "K-nearest neighbors",
		Construct: func() Classifier {
			return &Neighbors{}
		},
	},
	"stumps": ClassifierDesc{
		Desc: "boosted tree stumps",
		Construct: func() Classifier {
			return &Stumps{}
		},
	},
}
