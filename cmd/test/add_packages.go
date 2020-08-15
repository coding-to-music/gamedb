package main

import (
	"bufio"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

//noinspection GoUnusedFunction
func addPackages() {

	file, err := os.Open("ids.txt")
	if err != nil {
		zap.S().Error(err)
		return
	}
	defer func() {
		err = file.Close()
		zap.S().Error(err)
	}()

	var (
		wg     sync.WaitGroup
		locked bool
	)

	go func() {
		for {
			time.Sleep(time.Second * 5)

			c, err := queue.ProducerChannels[queue.QueuePackages].Inspect()
			if err != nil {
				zap.S().Error(err)
				continue
			}

			if c.Messages >= 10 && !locked {
				locked = true
				wg.Add(1)
				zap.S().Info(time.Now().Format(helpers.DateSQL), "locked")
			} else if c.Messages < 10 && locked {
				locked = false
				wg.Done()
				zap.S().Info(time.Now().Format(helpers.DateSQL), "unlocked")
			}
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		wg.Wait()

		packageID, err := strconv.Atoi(scanner.Text())
		if err != nil {
			zap.S().Error(err)
			continue
		}

		if packageID < 236_400 {
			continue
		}

		zap.S().Info(time.Now().Format(helpers.DateSQL), packageID)

		pack := mongo.Package{}

		err = mongo.FindOne(mongo.CollectionPackages, bson.D{{"_id", packageID}}, nil, bson.M{"_id": 1}, &pack)
		if err != nil && err != mongo.ErrNoDocuments {
			zap.S().Error(err)
			continue
		}

		if err == mongo.ErrNoDocuments {
			err = queue.ProduceSteam(queue.SteamMessage{PackageIDs: []int{packageID}})
			if err != nil {
				zap.S().Error(err)
			} else {
				// Success
				time.Sleep(time.Second)
			}
		}
	}

	if scanner.Err() != nil {
		zap.S().Error(scanner.Err())
	}
}
