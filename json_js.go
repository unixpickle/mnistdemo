// +build js

package mnistdemo

import "github.com/gopherjs/gopherjs/js"

func unmarshalTrees(data []byte, output *[]*archivedTree) error {
	jsObj := js.Global.Get("JSON").Call("parse", string(data))
	var res []*archivedTree
	for i := 0; i < jsObj.Length(); i++ {
		treeObj := jsObj.Index(i)
		res = append(res, jsObjToTree(treeObj))
	}
	*output = res
	return nil
}

func jsObjToTree(obj *js.Object) *archivedTree {
	if c := obj.Get("Classification"); c != nil {
		res := &archivedTree{
			Classification: map[string]float64{},
		}
		for _, key := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
			if c.Call("hasOwnProperty", key).Bool() {
				res.Classification[key] = c.Get(key).Float()
			}
		}
		return res
	}
	return &archivedTree{
		Pixel:     obj.Get("Pixel").Int(),
		Threshold: obj.Get("Threshold").Float(),
		LessEqual: jsObjToTree(obj.Get("LessEqual")),
		Greater:   jsObjToTree(obj.Get("Greater")),
	}
}
