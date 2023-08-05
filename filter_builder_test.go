package go_mongo_repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/xompass/lbq"
)

type JSONTest struct {
	path string
	json string
}

var tests = []JSONTest{
	{path: "test_files/empty.json"},
	{path: "test_files/nested_query.json"},
	{path: "test_files/inq.json"},
	{path: "test_files/and.json"},
	{path: "test_files/complex.json"},
	{path: "test_files/medium.json"},
}

func TestLbFilterToBson(t *testing.T) {
	repository, err := NewRepository[AssetTest](nil, RepositoryOptions{Created: true, Modified: true, Deleted: true})
	if err != nil {
		t.Fatal(err)
	}
	for i, testFile := range tests {
		fmt.Println(testFile.path)
		content, _ := os.ReadFile(testFile.path)

		lbFilter, err := lbq.ParseFilter(string(content))
		if err != nil {
			fmt.Println(err)
		}

		query, err := lbFilterQuery(*lbFilter, repository.schema)
		if i == 1 {
			if err == nil {
				t.Fatal(errors.New("must be error"))
			} else {
				if err.Error() != "can not query on nested fields" {
					t.Fatal("invalid error")
				}
			}
		} else if err != nil {
			t.Fatalf(err.Error())
		}

		_json, _ := json.MarshalIndent(query, "", "\t")
		fmt.Println(string(_json))
	}
}

func BenchmarkLbFilterToBson(b *testing.B) {
	repository, _ := NewRepository[AssetTest](nil, RepositoryOptions{Created: true, Modified: true, Deleted: true})
	var filters []lbq.Filter
	for _, testFile := range tests {
		content, _ := os.ReadFile(testFile.path)

		testFile.json = string(content)
		filter, _ := lbq.ParseFilter(testFile.json)
		filters = append(filters, *filter)
	}

	for ti, jsonTest := range tests {
		b.Run(jsonTest.path, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = lbFilterQuery(filters[ti], repository.schema)
			}
		})
	}
}
