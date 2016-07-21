package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/unixpickle/mnist"
	"github.com/unixpickle/mnist-demo"
	"github.com/unixpickle/serializer"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <classifier> <output_file>\n", os.Args[0])
		printClassifiers()
		os.Exit(1)
	}

	desc, ok := mnistdemo.Classifiers[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown classifier:", os.Args[1])
		os.Exit(1)
	}

	classifier := desc.Construct()
	classifier.Train(mnistSamples(mnist.LoadTrainingDataSet()),
		mnistSamples(mnist.LoadTestingDataSet()))

	resData, err := serializer.SerializeWithType(classifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to serialize:", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(os.Args[2], resData, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save:", err)
		os.Exit(1)
	}
}

func printClassifiers() {
	var names []string
	for name := range mnistdemo.Classifiers {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprintln(os.Stderr, "\nAvailable classifiers:")
	for _, name := range names {
		desc := mnistdemo.Classifiers[name].Desc
		fmt.Fprintf(os.Stderr, " %s - %s\n", name, desc)
	}
	fmt.Fprintln(os.Stderr)
}

func mnistSamples(d mnist.DataSet) []*mnistdemo.TrainingSample {
	var res []*mnistdemo.TrainingSample
	for _, sample := range d.Samples {
		ts := &mnistdemo.TrainingSample{
			Label:  sample.Label,
			Sample: new(mnistdemo.Sample),
		}
		copy(ts.Sample[:], sample.Intensities)
		res = append(res, ts)
	}
	return res
}
