package utils

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

type queuePackages struct{}

func (queuePackages) name() string {
	return "queue-packages"
}

func (queuePackages) run() {

	file, err := os.Open("ids.txt")
	if err != nil {
		log.ErrS(err)
		return
	}

	defer helpers.Close(file)

	var (
		wg     sync.WaitGroup
		locked bool
	)

	go func() {
		for {
			time.Sleep(time.Second * 5)

			c, err := queue.ProducerChannels[queue.QueuePackages].Inspect()
			if err != nil {
				log.ErrS(err)
				continue
			}

			if c.Messages >= 10 && !locked {
				locked = true
				wg.Add(1)
				log.InfoS(time.Now().Format(helpers.DateSQL) + " locked")
			} else if c.Messages < 10 && locked {
				locked = false
				wg.Done()
				log.InfoS(time.Now().Format(helpers.DateSQL) + " unlocked")
			}
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		wg.Wait()

		packageID, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.ErrS(err)
			continue
		}

		if packageID < 236_400 {
			continue
		}

		log.Info(time.Now().Format(helpers.DateSQL), zap.Int("package", packageID))

		pack := mongo.Package{}

		err = mongo.FindOne(mongo.CollectionPackages, bson.D{{"_id", packageID}}, nil, bson.M{"_id": 1}, &pack)
		if err != nil && err != mongo.ErrNoDocuments {
			log.ErrS(err)
			continue
		}

		if err == mongo.ErrNoDocuments {
			err = queue.ProduceSteam(queue.SteamMessage{PackageIDs: []int{packageID}})
			if err != nil {
				log.ErrS(err)
			} else {
				// Success
				time.Sleep(time.Second)
			}
		}
	}

	if scanner.Err() != nil {
		log.ErrS(scanner.Err())
	}
}
