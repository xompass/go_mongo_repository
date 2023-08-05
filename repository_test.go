package go_mongo_repository

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xompass/lbq"
)

func initializeDataSource() (*MongoDatasource, error) {
	mongoDatasource := MongoDatasource{}
	opts := MongoConnectorOpts{
		Database: "xompass_cloud_db",
	}

	opts.ClientOptions.ApplyURI("go_mongo_repository://localhost:27017")

	_, err := mongoDatasource.NewConnector("db", opts)
	if err != nil {
		return nil, err
	}

	return &mongoDatasource, nil
}

func createRepository() (*MongoRepository[AssetTest], error) {
	mDatasource, err := initializeDataSource()
	if err != nil {
		return nil, err
	}

	return NewRepository[AssetTest](mDatasource, RepositoryOptions{Created: true, Modified: true, Deleted: true})
}

func TestMongoRepository(t *testing.T) {
	repository, err := createRepository()

	if err != nil {
		t.Fatal(err)
	}

	var asset *AssetTest
	var assets []AssetTest

	count, err := repository.Count(lbq.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(count)

	asset, err = repository.FindOne(lbq.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	printJSON(asset)

	asset, err = repository.FindById(asset.Id, lbq.Filter{Fields: map[string]bool{"name": true, "id": true}})
	if err != nil {
		t.Fatal(err)
	}
	printJSON(asset)

	err = repository.DeleteById(asset.Id)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := repository.Exists(asset.Id)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(exists)

	assets, err = repository.Find(lbq.Filter{
		Limit:  2,
		Skip:   1,
		Fields: map[string]bool{"name": true, "id": true},
		Order:  []lbq.Order{{Field: "name", Direction: "ASC"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	printJSON(assets)
}

func printJSON(content interface{}) {
	_json, _ := json.Marshal(content)
	fmt.Println(string(_json))
}
