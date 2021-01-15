package tasks

import (
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
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
		return err
	}

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		message := &generated.GroupsRequest{
			Pagination: &generated.PaginationRequest{
				Offset: offset,
				Limit:  limit,
			},
			Projection: []string{"_id", "type", "primaries"},
		}

		resp, err := generated.NewGroupsServiceClient(conn).List(ctx, message)
		if err != nil {
			return err
		}

		groups := resp.GetGroups()
		for _, group := range groups {

			err = queue.ProduceGroupPrimaries(group.GetID(), helpers.GroupTypeGroup, int(group.GetPrimaries()))
			if err != nil {
				return err
			}
		}

		if int64(len(groups)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
