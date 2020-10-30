package main

import (
	"context"

	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/golang/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupsServer struct {
	generated.UnimplementedGroupsServiceServer
}

func (g GroupsServer) Stream(request *generated.GroupsRequest, response generated.GroupsService_StreamServer) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000
	var projection = helpers.MakeMongoProjection(request.GetProjection())
	var filter = g.makeFilter(request)

	for {

		groups, err := mongo.GetGroups(limit, offset, bson.D{{"_id", 1}}, filter, projection)
		if err != nil {
			return err
		}

		for _, group := range groups {

			err = response.Send(g.makeGroup(group))
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

func (g GroupsServer) List(ctx context.Context, request *generated.GroupsRequest) (response *generated.GroupsResponse, err error) {

	sort := helpers.MakeMongoOrder(request.GetPagination())
	projection := helpers.MakeMongoProjection(request.GetProjection())
	filter := g.makeFilter(request)

	groups, err := mongo.GetGroups(request.GetPagination().GetOffset(), request.GetPagination().GetLimit(), sort, filter, projection)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	if err != nil {
		return nil, err
	}

	filtered, err := mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.GroupsResponse{}
	response.Pagination = helpers.MakePaginationResponse(request.GetPagination(), total, filtered)

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
		Type:          m.Type,
		Primaries:     int32(m.Primaries),
	}
}

func (g GroupsServer) makeFilter(request *generated.GroupsRequest) (b bson.D) {

	filter := bson.D{}

	if len(request.GetIDs()) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": request.GetIDs()}})
	}

	if len(request.GetType()) > 0 {
		filter = append(filter, bson.E{Key: "type", Value: request.GetType()})
	}

	return b
}
