package utils

import (
	"arrayexpress-fetch/dtos"
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func WriteMetadata(timestamp *sync.Map, status *sync.Map, mongoClient *mongo.Client) {
	current_time := time.Now().UnixMilli()

	status.Range(func(key, value interface{}) bool {
		dataModifiedAt, ok := timestamp.Load(key)

		body := dtos.AccessionLogs{
			Accession: key.(string),
			Status:    value.(string),
			FetchedAt: current_time,
		}

		if ok {
			body.ModifiedAt = dataModifiedAt.(int64)
		}

		_, err := mongoClient.Database("arrayexpress").Collection("accession").UpdateOne(context.Background(), bson.M{
			"accession": key,
		}, bson.M{"$set": body}, options.Update().SetUpsert(true))

		if err != nil {
			fmt.Println("Write Failed: ", err)
		}

		return true
	})
}

func ReadMetadata(mongoClient *mongo.Client) *sync.Map {
	cur, err := mongoClient.Database("arrayexpress").Collection("accession").Find(context.TODO(), bson.M{})

	timestamps := sync.Map{}

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return &timestamps
	}

	for cur.Next(context.Background()) {
		var log dtos.AccessionLogs
		err := cur.Decode(&log)

		if err != nil {
			fmt.Println("Read Failed: ", err)
			continue
		}

		timestamps.Store(log.Accession, int64(log.ModifiedAt))
	}

	return &timestamps
}
