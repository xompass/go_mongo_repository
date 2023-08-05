package go_mongo_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xompass/lbq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IModel interface {
	GetTableName() string
	GetModelName() string
	GetPluralModelName() string
	GetConnectorName() string
	GetId() interface{}
}

type FieldDetails struct {
	BsonName  string
	JsonName  string
	FieldType string
}

type MongoRepository[T IModel] struct {
	Options    RepositoryOptions
	collection *mongo.Collection
	schema     *Schema
	connector  *MongoConnector
}

type RepositoryOptions struct {
	Created  bool
	Modified bool
	Deleted  bool
}

type UpdateOptions struct {
	Insert bool
	Update bool
}

func NewRepository[T IModel](ds *MongoDatasource, options RepositoryOptions) (*MongoRepository[T], error) {
	instance := *new(T)
	collectionName := instance.GetTableName()

	schema := NewSchema(instance)

	err := ds.RegisterModel(instance)
	if err != nil {
		return nil, err
	}

	connector, _ := ds.GetModelConnector(instance)
	if connector == nil {
		return &MongoRepository[T]{
			Options:    options,
			collection: nil,
			schema:     schema,
			connector:  nil,
		}, nil
	}

	connectorOpts := connector.GetOptions()
	client := connector.GetDriver()

	databaseName := connectorOpts.Database
	if databaseName == "" {
		return nil, errors.New("database name is required")
	}

	repository := &MongoRepository[T]{
		Options:    options,
		collection: client.Database(databaseName).Collection(collectionName),
		schema:     schema,
		connector:  connector,
	}

	return repository, nil
}

func (repository *MongoRepository[T]) GetCollection() *mongo.Collection {
	return repository.collection
}

func (repository *MongoRepository[T]) GetSchema() *Schema {
	return repository.schema
}

func (repository *MongoRepository[T]) Find(filter lbq.Filter) ([]T, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := repository.fixQuery(parsedFilter.Where)

	cursor, err := repository.collection.Find(ctx, query, &options.FindOptions{
		Sort:       parsedFilter.Options.Sort,
		Limit:      parsedFilter.Options.Limit,
		Skip:       parsedFilter.Options.Skip,
		Projection: parsedFilter.Options.Fields,
	})

	if err != nil {
		return nil, err
	}

	var receiver []T
	if err = cursor.All(ctx, &receiver); err != nil {
		return nil, err
	}

	if receiver == nil {
		return []T{}, nil
	}
	return receiver, nil
}

func (repository *MongoRepository[T]) FindOne(filter lbq.Filter) (*T, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return nil, err
	}
	receiver := new(T)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := repository.fixQuery(parsedFilter.Where)
	err = repository.collection.FindOne(ctx, query, &options.FindOneOptions{
		Sort:       parsedFilter.Options.Sort,
		Skip:       parsedFilter.Options.Skip,
		Projection: parsedFilter.Options.Fields,
	}).Decode(receiver)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return receiver, err
}

func (repository *MongoRepository[T]) FindById(id interface{}, filter lbq.Filter) (*T, error) {
	if filter.Where == nil || len(filter.Where) == 0 {
		filter.Where = lbq.Where{"id": id}
	} else {
		filter.Where = lbq.Where{
			"and": lbq.AndOrCondition{
				lbq.Where{"id": id},
				filter.Where,
			},
		}
	}

	return repository.FindOne(filter)
}

func (repository *MongoRepository[T]) Insert(doc T) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	document, err := repository.fixInsert(doc)
	if err != nil {
		return nil, err
	}

	insertedResult, err := repository.collection.InsertOne(ctx, document)

	if err != nil {
		return nil, err
	}

	return insertedResult.InsertedID, nil
}

func (repository *MongoRepository[T]) Create(doc T) (*T, error) {
	insertedID, err := repository.Insert(doc)
	if err != nil {
		return nil, err
	}

	return repository.FindById(insertedID, lbq.Filter{})
}

func (repository *MongoRepository[T]) FindOneOrCreate(filter lbq.Filter, doc T) (*T, error) {
	upsert := true
	after := options.After

	return repository.findOneAnUpdate(filter, doc, &options.FindOneAndUpdateOptions{Upsert: &upsert, ReturnDocument: &after})
}

func (repository *MongoRepository[T]) Upsert(filter lbq.Filter, update any) error {
	upsert := true
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fixedUpdate, err := repository.fixUpdate(update, UpdateOptions{}, UpdateOptions{})
	if err != nil {
		return err
	}

	query := repository.fixQuery(parsedFilter.Where)

	_, err = repository.collection.UpdateOne(ctx, query, fixedUpdate, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return err
	}

	return nil
}

func (repository *MongoRepository[T]) UpdateOne(filter lbq.Filter, update interface{}) error {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fixedUpdate, err := repository.fixUpdate(update, UpdateOptions{}, UpdateOptions{})
	if err != nil {
		return err
	}

	query := repository.fixQuery(parsedFilter.Where)

	_, err = repository.collection.UpdateOne(ctx, query, fixedUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (repository *MongoRepository[T]) UpdateById(id interface{}, update interface{}) error {
	return repository.UpdateOne(lbq.Filter{
		Where: lbq.Where{"id": id},
	}, update)
}

func (repository *MongoRepository[T]) FindOneAnUpdate(filter lbq.Filter, update interface{}) (*T, error) {
	return repository.findOneAnUpdate(filter, update)
}

func (repository *MongoRepository[T]) findOneAnUpdate(filter lbq.Filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) (*T, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var updateOptions *options.FindOneAndUpdateOptions
	setCreated := false
	if len(opts) > 0 {
		updateOptions = opts[0]
		setCreated = *updateOptions.Upsert
	} else {
		updateOptions = &options.FindOneAndUpdateOptions{}
	}

	updateOptions.Projection = filter.Fields
	if updateOptions.ReturnDocument == nil {
		afterUpdate := options.After
		updateOptions.ReturnDocument = &afterUpdate
	}

	fixedUpdate, err := repository.fixUpdate(update, UpdateOptions{}, UpdateOptions{Insert: setCreated})
	if err != nil {
		return nil, err
	}

	query := repository.fixQuery(parsedFilter.Where)

	receiver := new(T)
	err = repository.collection.FindOneAndUpdate(ctx, query, fixedUpdate, updateOptions).Decode(receiver)

	fmt.Println(receiver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return receiver, err
}

func (repository *MongoRepository[T]) UpdateMany(filter lbq.Filter, update interface{}) (int64, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fixedUpdate, err := repository.fixUpdate(update, UpdateOptions{}, UpdateOptions{})
	if err != nil {
		return 0, err
	}

	query := repository.fixQuery(parsedFilter.Where)

	result, err := repository.collection.UpdateMany(ctx, query, fixedUpdate)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (repository *MongoRepository[T]) Count(filter lbq.Filter) (int64, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := repository.fixQuery(parsedFilter.Where)

	return repository.collection.CountDocuments(ctx, query)
}

func (repository *MongoRepository[T]) Exists(id interface{}) (bool, error) {
	doc, err := repository.FindOne(lbq.Filter{
		Where: lbq.Where{"id": id},
		Fields: map[string]bool{
			"_id": true,
		},
	})
	if err != nil {
		return false, err
	}

	if doc != nil {
		return true, nil
	}

	return false, nil
}

func (repository *MongoRepository[T]) DeleteOne(filter lbq.Filter) error {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := repository.fixQuery(parsedFilter.Where)
	if repository.Options.Deleted {
		result, err := repository.collection.UpdateOne(ctx, query, bson.M{"$currentDate": bson.M{"deleted": true}})
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return errors.New("no documents founds")
		}
		return nil
	}

	result, err := repository.collection.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("no documents founds")
	}

	return nil
}

func (repository *MongoRepository[T]) DeleteById(id interface{}) error {
	return repository.DeleteOne(lbq.Filter{
		Where: lbq.Where{"id": id},
	})
}

func (repository *MongoRepository[T]) DeleteMany(filter lbq.Filter) (int64, error) {
	parsedFilter, err := lbFilterQuery(filter, repository.schema)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := repository.fixQuery(parsedFilter.Where)
	if repository.Options.Deleted {
		result, err := repository.collection.UpdateMany(ctx, query, bson.M{"$currentDate": bson.M{"deleted": true}})
		if err != nil {
			return 0, err
		}
		return result.ModifiedCount, nil
	}

	result, err := repository.collection.DeleteMany(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (repository *MongoRepository[T]) fixQuery(query bson.M) bson.M {
	if repository.Options.Deleted {
		query = getSoftDeleteQuery(query)
	}

	return query
}

func (repository *MongoRepository[T]) fixUpdate(update interface{}, updateDeleted UpdateOptions, setCreated UpdateOptions) (bson.M, error) {
	document, err := toBsonMap(update)
	if err != nil {
		return nil, err
	}

	hasFields := false
	hasCommands := false
	for key := range document {
		if strings.HasPrefix(key, "$") {
			hasCommands = true
		} else {
			hasFields = true
		}
	}

	if hasFields && hasCommands {
		return bson.M{}, errors.New("the update has a mix between fields and commands")
	}

	var newUpdate bson.M
	var bsonSet bson.M

	if hasCommands {
		set, ok := document["$set"]
		if ok {
			newUpdate = document
			bsonSet, _ = set.(bson.M)
		} else {
			newUpdate = bson.M{}
			bsonSet = bson.M{}
		}
	}

	if hasFields {
		newUpdate = bson.M{}
		bsonSet = document
	}

	// Remove created, deleted and modified fields from update. This is managed by the repository
	if repository.Options.Created {
		delete(bsonSet, "created")
	}

	if repository.Options.Modified {
		delete(bsonSet, "modified")
	}

	if repository.Options.Deleted {
		delete(bsonSet, "deleted")
	}

	if len(bsonSet) > 0 {
		newUpdate["$set"] = bsonSet
	} else {
		delete(bsonSet, "$set")
	}

	if repository.Options.Modified || repository.Options.Created || repository.Options.Deleted {
		currentDate, ok := document["$currentDate"]
		var bsonCurrentDate bson.M
		if ok {
			bsonCurrentDate, _ = currentDate.(bson.M)
		} else {
			bsonCurrentDate = bson.M{}
		}

		// The "modified" date is set
		if repository.Options.Modified {
			bsonCurrentDate["modified"] = true
		}

		if repository.Options.Deleted {
			// The "deleted" date is set if required
			if updateDeleted.Update {
				bsonCurrentDate["deleted"] = true
			} else {
				delete(bsonCurrentDate, "deleted")
			}
		}

		if repository.Options.Created {
			// The "created" date is set if required
			if setCreated.Update && !setCreated.Insert {
				bsonCurrentDate["created"] = true
			} else {
				delete(bsonCurrentDate, "created")
			}
		}

		if len(bsonCurrentDate) > 0 {
			newUpdate["$currentDate"] = bsonCurrentDate
		} else {
			delete(newUpdate, "$currentDate")
		}
	}

	if repository.Options.Created && setCreated.Insert {
		temp, ok := newUpdate["$setOnInsert"]
		var setOnInsert bson.M
		if ok {
			setOnInsert, _ = temp.(bson.M)
		} else {
			setOnInsert = bson.M{}
		}

		// The "created" date is set on insert if required
		setOnInsert["created"] = time.Now()
		newUpdate["$setOnInsert"] = setOnInsert
	}

	return newUpdate, nil
}

func (repository *MongoRepository[T]) fixInsert(doc interface{}) (bson.M, error) {
	document, err := toBsonMap(doc)
	if err != nil {
		return nil, err
	}

	if repository.Options.Created {
		document["created"] = time.Now()
	}

	if repository.Options.Modified {
		document["modified"] = time.Now()
	}

	if repository.Options.Deleted {
		document["deleted"] = nil
	}

	return document, nil
}

func getSoftDeleteQuery(query bson.M) bson.M {
	return bson.M{
		"$and": []interface{}{
			query,
			bson.M{"deleted": bson.M{"$type": 10}},
		},
	}
}

func toBsonMap(v interface{}) (doc bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return doc, err
}
