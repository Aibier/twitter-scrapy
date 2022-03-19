package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TwitPost struct {
	Id       primitive.ObjectID `json:"id,omitempty"`
	Title    string             `json:"text,omitempty" validate:"required"`
	CreatedAt time.Time
}
