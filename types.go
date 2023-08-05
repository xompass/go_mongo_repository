package go_mongo_repository

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"time"
)

type MongoDate struct {
	time.Time
}

var dateFormat = "2006-01-02T15:04:05.000Z"

func (date *MongoDate) UnmarshalBSON(data []byte) error {
	dateReader := bsonrw.NewBSONValueReader(bsontype.DateTime, data)
	milliseconds, err := dateReader.ReadDateTime()
	if err != nil {
		fmt.Println("error", err)
		return err
	}

	*date = MongoDate{time.Unix(0, milliseconds*int64(time.Millisecond))}
	return nil
}

func (date *MongoDate) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", date.Time.Format(dateFormat))
	return []byte(stamp), nil
}
