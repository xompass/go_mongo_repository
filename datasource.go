package go_mongo_repository

import (
	"errors"
	"fmt"
)

type MongoDatasource struct {
	connectors           map[string]*MongoConnector
	connectorByModelName map[string]*MongoConnector
}

func (receiver *MongoDatasource) NewConnector(name string, clientOptions MongoConnectorOpts) (*MongoDatasource, error) {
	clientOptions.Name = name
	if clientOptions.Database == "" {
		return nil, errors.New("database value is required")
	}
	connector, err := NewMongoConnector(&clientOptions)
	if err != nil {
		return nil, err
	}
	if receiver.connectors == nil {
		receiver.connectors = make(map[string]*MongoConnector)
	}
	receiver.connectors[name] = connector
	return receiver, nil
}

func (receiver *MongoDatasource) Destroy() {
	for _, connector := range receiver.connectors {
		_ = connector.Disconnect()
	}
}

func (receiver *MongoDatasource) RegisterModel(model IModel) error {
	connectorName := model.GetConnectorName()
	modelName := model.GetModelName()
	connector, err := receiver.GetConnector(connectorName)
	if err != nil {
		return err
	}
	if receiver.connectorByModelName == nil {
		receiver.connectorByModelName = make(map[string]*MongoConnector)
	}

	receiver.connectorByModelName[modelName] = connector
	return nil
}

func (receiver *MongoDatasource) GetModelConnector(model IModel) (*MongoConnector, error) {
	connector, ok := receiver.connectorByModelName[model.GetModelName()]
	if !ok {
		return nil, fmt.Errorf("the model %s is not registered", model.GetModelName())
	}

	return connector, nil
}

func (receiver *MongoDatasource) GetConnector(name string) (*MongoConnector, error) {
	connector, ok := receiver.connectors[name]
	if !ok {
		return nil, fmt.Errorf("connector with name %s does not exists", name)
	}
	return connector, nil
}
