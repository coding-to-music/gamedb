package helpers

import (
	"go.mongodb.org/mongo-driver/bson"
)

// Just here so frontend can use it
var LastUpdatedQuery = bson.D{
	{"community_visibility_state", 3}, // Public
	{"removed", false},
}
