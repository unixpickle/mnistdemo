// +build !js

package mnistdemo

import "encoding/json"

func unmarshalTrees(data []byte, output *[]*archivedTree) error {
	return json.Unmarshal(data, output)
}
