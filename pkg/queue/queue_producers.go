package queue

import (
	"errors"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
)

var ErrInQueue = errors.New("already in queue")

type SteamPayload struct {
	AppIDs     []int
	PackageIDs []int
	ProfileIDs []int64
	Force      bool
}

func ProduceToSteam(payload SteamPayload) (err error) {

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 && len(payload.ProfileIDs) == 0 {
		return nil
	}

	time.Sleep(time.Millisecond)

	var appIDs []int
	var packageIDs []int
	var profileIDs []int64

	mc := memcache.GetClient()

	for _, appID := range payload.AppIDs {

		if !config.IsLocal() && !payload.Force {

			item := memcache.MemcacheAppInQueue(appID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.AppIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		appIDs = append(appIDs, appID)
	}

	for _, packageID := range payload.PackageIDs {

		if !config.IsLocal() && !payload.Force {

			item := memcache.MemcachePackageInQueue(packageID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.PackageIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		packageIDs = append(packageIDs, packageID)
	}

	for _, profileID := range payload.ProfileIDs {

		if !config.IsLocal() && !payload.Force {

			item := memcache.MemcachePlayerInQueue(profileID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.ProfileIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		profileIDs = append(profileIDs, profileID)
	}

	message := &steamMessage{
		Message: steamMessageInner{
			AppIDs:     appIDs,
			PackageIDs: packageIDs,
			PlayerIDs:  profileIDs,
		},
	}
	message.Force = payload.Force

	return produce(message, QueueSteam)
}

func ProduceGroup(ids []string, force bool) (err error) {

	if len(ids) == 0 {
		return nil
	}

	time.Sleep(time.Millisecond)

	mc := memcache.GetClient()

	var filteredIDs []string

	for _, id := range ids {

		id = strings.TrimSpace(id)

		if helpers.IsValidGroupID(id) {

			if !config.IsLocal() && !force {

				item := memcache.MemcacheGroupInQueue(id)

				_, err := mc.Get(item.Key)
				if err == nil {
					continue
				}

				err = mc.Set(&item)
				log.Err(err)
			}

			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) == 0 {
		return nil
	}

	chunks := helpers.ChunkStrings(filteredIDs, 10)

	for _, chunk := range chunks {

		message := &groupMessage{
			Message: groupMessageInner{
				IDs: chunk,
			},
		}
		message.Force = force

		err = produce(message, queueGroups)
		log.Err(err)
	}

	return nil
}
