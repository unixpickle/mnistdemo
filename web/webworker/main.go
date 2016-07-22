package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mnistdemo"
	"github.com/unixpickle/serializer"
)

var Classifier mnistdemo.Classifier

func main() {
	js.Global.Set("onmessage", js.MakeFunc(messageHandler))
}

func messageHandler(this *js.Object, dataArg []*js.Object) interface{} {
	if len(dataArg) != 1 {
		panic("expected one argument")
	}
	data := dataArg[0].Get("data")
	command := data.Index(0).String()

	switch command {
	case "init":
		initCommand(data.Index(1).Interface().([]byte))
	case "classify":
		classifyCommand(data.Index(1))
	}

	return nil
}

func initCommand(data []byte) {
	c, err := serializer.DeserializeWithType(data)
	if err != nil {
		emitInitialized(err.Error())
		return
	}
	var ok bool
	Classifier, ok = c.(mnistdemo.Classifier)
	if !ok {
		emitInitialized("invalid datatype")
		return
	}
	emitInitialized(nil)
}

func classifyCommand(obj *js.Object) {
	sample := new(mnistdemo.Sample)
	for i := 0; i < 28*28; i++ {
		sample[i] = obj.Index(i).Float()
	}
	class := Classifier.Classify(sample)
	emitClassification(class)
}

func emitInitialized(err interface{}) {
	js.Global.Call("postMessage", map[string]interface{}{"init": err})
}

func emitClassification(class int) {
	js.Global.Call("postMessage", map[string]int{"classification": class})
}
