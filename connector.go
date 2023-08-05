package go_mongo_repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnectorOpts struct {
	options.ClientOptions
	Name     string
	Database string
}

type MongoConnector struct {
	ctx     context.Context
	client  *mongo.Client
	options *MongoConnectorOpts
}

func NewMongoConnector(opts *MongoConnectorOpts) (*MongoConnector, error) {
	ctx := context.Background()
	connector := &MongoConnector{
		ctx:     ctx,
		options: opts,
	}

	err := connector.connect()
	if err != nil {
		return nil, err
	}

	if err := connector.ping(); err != nil {
		return nil, err
	}

	return connector, nil
}

func (receiver *MongoConnector) connect() error {
	opts := receiver.options.ClientOptions
	client, err := mongo.Connect(receiver.ctx, &opts)

	if err != nil {
		return err
	}

	receiver.client = client
	return nil
}

func (receiver *MongoConnector) ping() error {
	if receiver.client == nil {
		return errors.New("go_mongo_repository client not initialized")
	}
	return receiver.client.Ping(receiver.ctx, nil)
}

func (receiver *MongoConnector) Disconnect() error {
	if receiver.client == nil {
		return errors.New("go_mongo_repository client not initialized")
	}
	return receiver.client.Disconnect(receiver.ctx)
}

func (receiver *MongoConnector) GetDriver() *mongo.Client {
	return receiver.client
}

func (receiver *MongoConnector) GetOptions() MongoConnectorOpts {
	return *receiver.options
}
