package util

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func PatchGen(current, expected, kind interface{}) ([]byte, error) {
	currentJson, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}
	expectedJson, _ := json.Marshal(expected)
	if err != nil {
		return nil, err
	}
	patchBytes, _ := strategicpatch.CreateTwoWayMergePatch(currentJson, expectedJson, kind)
	return patchBytes, nil
}
