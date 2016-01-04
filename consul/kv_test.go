package consul

import (
	"testing"
)

const ValidKey = "/_wakeful/nodes/bar"
const Base64Value = "Zm9vYmFy" // => foobar

var kv = ConsulKV{Key: ValidKey, Value: Base64Value}

func TestKVValueDecoding(t *testing.T) {
	value := kv.DecodedValue()

	expectedValue := "bar"
	if value != expectedName {
		t.Errorf("expect: %s but got: %s", expectedValue, value)
	}
}
