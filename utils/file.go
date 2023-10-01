package utils

import (
	"arrayexpress-fetch/constants"
	"arrayexpress-fetch/dtos"
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func WriteMetadata(metadata map[string][]dtos.ResultMetadata) {
	fp, err := os.OpenFile(fmt.Sprintf("%smetadata.txt", constants.FILE_BASE_PATH), os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return
	}

	defer fp.Close()

	for key, value := range metadata {
		fp.WriteString(fmt.Sprintf("%s: ", key))

		for _, accession := range value {
			fp.WriteString(fmt.Sprintf("%s,", accession.Name))
		}

		fp.WriteString("\n")
	}
}

func WriteTimestamp(timestamp map[string]int64, mongoClient *mongo.Client) {
	current_time := time.Now().UnixMilli()

	for key, value := range timestamp {
		_, err := mongoClient.Database("arrayexpress").Collection("timestamp").UpdateOne(context.Background(), bson.M{
			"accession": key,
		}, bson.M{
			"$set": dtos.AccessionLogs{
				Accession:  key,
				ModifiedAt: value,
				FetchedAt:  current_time,
			}}, options.Update().SetUpsert(true))

		if err != nil {
			fmt.Println("Write Failed: ", err)
			return
		}
	}
}

func ReadTimestamp(mongoClient *mongo.Client) map[string]int64 {
	cur, err := mongoClient.Database("arrayexpress").Collection("timestamp").Find(context.TODO(), bson.M{})

	timestamps := make(map[string]int64)

	if err != nil {
		fmt.Println("Read Failed: ", err)
		return timestamps
	}

	for cur.Next(context.Background()) {
		var log dtos.AccessionLogs
		err := cur.Decode(&log)

		if err != nil {
			fmt.Println("Read Failed: ", err)
			continue
		}

		timestamps[log.Accession] = int64(log.ModifiedAt)
	}

	return timestamps
}
