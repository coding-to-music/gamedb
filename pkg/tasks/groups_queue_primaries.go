package tasks

import (
	"io"

	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

type GroupsQueuePrimaries struct {
	BaseTask
}

func (c GroupsQueuePrimaries) ID() string {
	return "groups-queue-primaries"
}

func (c GroupsQueuePrimaries) Name() string {
	return "Queue all group primaries to be updated"
}

func (c GroupsQueuePrimaries) Group() TaskGroup {
	return TaskGroupGroups
}

func (c GroupsQueuePrimaries) Cron() TaskTime {
	return ""
}

func (c GroupsQueuePrimaries) work() (err error) {

	conn, ctx, err := backend.GetClient()
	if err != nil {
		log.Err(err.Error())
		return
	}

	message := &generated.GroupsRequest{
		Projection: []string{"_id", "type", "primaries"},
	}

	resp, err := generated.NewGroupsServiceClient(conn).Stream(ctx, message)
	if err != nil {
		log.Err(err.Error())
		return
	}

	for {

		group, err := resp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Err(err.Error())
			continue
		}

		err = queue.ProduceGroupPrimaries(group.GetID(), helpers.GroupTypeGroup, int(group.GetPrimaries()))
		if err != nil {
			log.Err(err.Error())
			continue
		}
	}

	return nil
}
