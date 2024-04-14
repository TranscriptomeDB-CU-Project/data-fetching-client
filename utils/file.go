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
	BULK_SIZE := 10000
	write_model := make([]mongo.WriteModel, BULK_SIZE)
	idx := 0

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

		write_model[idx] = mongo.NewUpdateOneModel().SetFilter(bson.M{"accession": key}).SetUpdate(bson.M{"$set": body}).SetUpsert(true)
		idx++

		if idx == BULK_SIZE {
			_, err := mongoClient.Database("arrayexpress").Collection("accession").BulkWrite(context.Background(), write_model, options.BulkWrite().SetOrdered(false))
			if err != nil {
				fmt.Println("Write Failed: ", err)
			}
			idx = 0
		}

		return true
	})

	if idx > 0 {
		write_model = write_model[:idx]
		_, err := mongoClient.Database("arrayexpress").Collection("accession").BulkWrite(context.Background(), write_model, options.BulkWrite().SetOrdered(false))
		if err != nil {
			fmt.Println("Write Failed: ", err)
		}
	}
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
