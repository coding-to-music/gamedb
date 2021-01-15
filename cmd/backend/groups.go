package main

import (
	"context"

	backendHelpers "github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/golang/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupsServer struct {
	generated.UnimplementedGroupsServiceServer
}

func (g GroupsServer) List(ctx context.Context, request *generated.GroupsRequest) (response *generated.GroupsResponse, err error) {

	sort := backendHelpers.MakeMongoOrder(request.GetPagination())
	projection := backendHelpers.MakeMongoProjection(request.GetProjection())

	filter := bson.D{
		{Key: "type", Value: helpers.GroupTypeGroup},
	}

	if len(request.GetIDs()) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": request.GetIDs()}})
	}

	groups, err := mongo.GetGroups(request.GetPagination().GetOffset(), request.GetPagination().GetLimit(), sort, filter, projection)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionGroups, bson.D{{Key: "type", Value: helpers.GroupTypeGroup}}, 0)
	if err != nil {
		return nil, err
	}

	filtered, err := mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.GroupsResponse{}
	response.Pagination = backendHelpers.MakePaginationResponse(request.GetPagination(), total, filtered)

	for _, group := range groups {
		response.Groups = append(response.Groups, g.makeGroup(group))
	}

	return response, err
}

func (g GroupsServer) Retrieve(ctx context.Context, request *generated.GroupRequest) (*generated.GroupResponse, error) {
	panic("implement me")
}

func (g GroupsServer) makeGroup(m mongo.Group) (r *generated.GroupResponse) {

	createdAt, err := ptypes.TimestampProto(m.CreatedAt)
	if err != nil {
		log.Err(err.Error())
	}

	updatedAt, err := ptypes.TimestampProto(m.UpdatedAt)
	if err != nil {
		log.Err(err.Error())
	}

	return &generated.GroupResponse{
		ID:            m.ID,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Name:          m.Name,
		Abbr:          m.Abbr,
		URL:           m.URL,
		AppID:         int32(m.AppID),
		Headline:      m.Headline,
		Summary:       m.Summary,
		Icon:          m.Icon,
		Trending:      float32(m.Trending),
		Members:       int32(m.Members),
		MembersInChat: int32(m.MembersInChat),
		MembersInGame: int32(m.MembersInGame),
		MembersOnline: int32(m.MembersOnline),
		Error:         m.Error,
		Primaries:     int32(m.Primaries),
	}
}
