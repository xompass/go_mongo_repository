package go_mongo_repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type PersistedModelWithStringId struct {
	Id *string `bson:"_id,omitempty" json:"id,omitempty"`
}

type PersistedModelWithId struct {
	Id *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
}

type PersistedModelWithDates struct {
	CreatedAt  *time.Time `bson:"created,omitempty" json:"created,omitempty"`
	ModifiedAt *time.Time `bson:"modified,omitempty" json:"modified,omitempty"`
}

type PersistedModelWithDeleted struct {
	DeletedAt *time.Time `bson:"deleted,omitempty" json:"deleted,omitempty"`
}

type FeatureTest struct {
	PersistedModelWithId      `bson:",inline" json:",inline"`
	PersistedModelWithDates   `bson:",inline" json:",inline"`
	PersistedModelWithDeleted `bson:",inline" json:",inline"`

	Type       *string     `bson:"type,omitempty" json:"type,omitempty"`
	Properties interface{} `bson:"properties,omitempty" json:"properties,omitempty"`
	Geometry   *struct {
		Type        string        `bson:"type,omitempty" json:"type,omitempty"`
		Coordinates []interface{} `bson:"coordinates,omitempty" json:"coordinates,omitempty"`
	} `bson:"geometry,omitempty" json:"geometry,omitempty"`
	AssetId *primitive.ObjectID `bson:"assetId,omitempty" json:"assetId,omitempty"`
	LayerId *primitive.ObjectID `bson:"layerId,omitempty" json:"layerId,omitempty"`
}

type AssetConfigTest struct {
	PersistedModelWithStringId `bson:",inline" json:",inline"`
	PersistedModelWithDates    `bson:",inline" json:",inline"`
	PersistedModelWithDeleted  `bson:",inline" json:",inline"`

	Status          interface{}  `bson:"status,omitempty" json:"status,omitempty"`
	Settings        interface{}  `bson:"settings,omitempty" json:"settings,omitempty"`
	Address         *string      `bson:"address,omitempty" json:"address,omitempty"`
	Feature         *FeatureTest `bson:"feature,omitempty" json:"feature,omitempty"`
	DataTTL         *float64     `bson:"dataTTL,omitempty" json:"dataTTL,omitempty"`
	LastEdgeAgentId *string      `bson:"lastEdgeAgentId,omitempty" json:"lastEdgeAgentId,omitempty"`
}

type AssetTest struct {
	PersistedModelWithId      `bson:",inline" json:",inline"`
	PersistedModelWithDates   `bson:",inline" json:",inline"`
	PersistedModelWithDeleted `bson:",inline" json:",inline"`

	Type              *string             `bson:"type,omitempty" json:"type,omitempty"`
	Name              *string             `bson:"name,omitempty" json:"name,omitempty"`
	Icon              *string             `bson:"icon,omitempty" json:"icon,omitempty"`
	Description       *string             `bson:"description,omitempty" json:"description,omitempty"`
	ReferenceId       *string             `bson:"referenceId,omitempty" json:"referenceId,omitempty"`
	Uri               *string             `bson:"uri,omitempty" json:"uri,omitempty"`
	Path              []string            `bson:"path,omitempty" json:"path,omitempty"`
	Requested         *time.Time          `bson:"requested,omitempty" json:"requested,omitempty"`
	AssetId           *primitive.ObjectID `bson:"assetId,omitempty" json:"assetId,omitempty"`
	Config            *AssetConfigTest    `bson:"_config,omitempty" json:"_config,omitempty"`
	CurrentStateId    *primitive.ObjectID `bson:"currentStateId,omitempty" json:"currentStateId,omitempty"`
	TemplateId        *primitive.ObjectID `bson:"templateId,omitempty" json:"templateId,omitempty"`
	AssetWizardTypeId *primitive.ObjectID `bson:"assetWizardTypeId,omitempty" json:"assetWizardTypeId,omitempty"`
	CustomerId        *primitive.ObjectID `bson:"customerId,omitempty" json:"customerId,omitempty"`
	ProjectId         *primitive.ObjectID `bson:"projectId,omitempty" json:"projectId,omitempty"`
	Asset             *AssetTest          `bson:"-" json:"asset,omitempty"`
}

type Sensor struct {
	PersistedModelWithId      `bson:",inline" json:",inline"`
	PersistedModelWithDates   `bson:",inline" json:",inline"`
	PersistedModelWithDeleted `bson:",inline" json:",inline"`

	Type               *string             `bson:"type,omitempty" json:"type,omitempty"`
	Name               *string             `bson:"name,omitempty" json:"name,omitempty"`
	Description        *string             `bson:"description,omitempty" json:"description,omitempty"`
	RelativeId         *string             `bson:"relativeId,omitempty" json:"relativeId,omitempty"`
	Enabled            *bool               `bson:"enabled,omitempty" json:"enabled,omitempty"`
	Unit               *string             `bson:"unit,omitempty" json:"unit,omitempty"`
	Parameters         interface{}         `bson:"parameters,omitempty" json:"parameters,omitempty"`
	Triggers           interface{}         `bson:"triggers,omitempty" json:"triggers,omitempty"`
	ExtendedProperties interface{}         `bson:"extendedProperties,omitempty" json:"extendedProperties,omitempty"`
	AssetId            *primitive.ObjectID `bson:"assetId,omitempty" json:"assetId,omitempty"`
	TemplateId         *primitive.ObjectID `bson:"templateId,omitempty" json:"templateId,omitempty"`
} // @name Sensor

func (a AssetTest) GetModelName() string {
	return "Asset"
}
func (a AssetTest) GetPluralModelName() string {
	return "Assets"
}

func (a AssetTest) GetTableName() string {
	return "Asset"
}

func (a AssetTest) GetConnectorName() string {
	return "db"
}

func (a AssetTest) GetId() interface{} {
	if a.Id == nil {
		return nil
	}
	return *a.Id
}

func (a Sensor) GetModelName() string {
	return "Sensor"
}

func (a Sensor) GetTableName() string {
	return "Sensor"
}

func (a Sensor) GetPluralModelName() string {
	return "Sensors"
}

func (a Sensor) GetId() interface{} {
	if a.Id == nil {
		return nil
	}
	return *a.Id
}
