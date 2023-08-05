package go_mongo_repository

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewSchema(t *testing.T) {
	schema := NewSchema(AssetTest{})
	_json, _ := json.MarshalIndent(schema.Relations, "", "\t")
	fmt.Println(string(_json))
	/*_json, _ = json.MarshalIndent(schema.Fields, "", "\t")
	fmt.Println(string(_json))*/

}
