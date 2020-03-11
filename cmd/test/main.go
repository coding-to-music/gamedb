package main

import (
	"bufio"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {

	config.SetVersion("test")
	log.Initialise([]log.LogName{log.LogNameTest})
	queue.Init(queue.QueueSteamDefinitions, false)

	file, err := os.Open("ids.txt")
	if err != nil {
		log.Err(err)
		return
	}
	defer func() {
		err = file.Close()
		log.Err(err)
	}()

	var (
		wg     sync.WaitGroup
		locked bool
	)

	go func() {
		for {
			time.Sleep(time.Second * 5)

			c, err := queue.Channels[rabbit.Producer][queue.QueuePackages].Inspect()
			if err != nil {
				log.Err(err)
				continue
			}

			if c.Messages >= 10 && locked == false {
				locked = true
				wg.Add(1)
				log.Info(time.Now().String(), "locked")
			} else if c.Messages < 10 && locked == true {
				locked = false
				wg.Done()
				log.Info(time.Now().String(), "unlocked")
			}
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		wg.Wait()

		packageID, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Err(err)
			continue
		}

		if packageID < 188_000 {
			continue
		}

		log.Info(time.Now().String(), packageID)

		pack := mongo.Package{}

		err = mongo.FindOne(mongo.CollectionPackages, bson.D{{"_id", packageID}}, nil, bson.M{"_id": 1}, &pack)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Err(err)
			continue
		}

		if err == mongo.ErrNoDocuments {
			err = queue.ProduceSteam(queue.SteamMessage{PackageIDs: []int{packageID}})
			if err != nil {
				log.Err(err)
			} else {
				// Success
				time.Sleep(time.Second)
			}
		}
	}

	if scanner.Err() != nil {
		log.Err(scanner.Err())
	}
}
