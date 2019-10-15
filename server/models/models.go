package models

import (
	"time"
)

type Model struct {
	CreatedOn time.Time  `json:"created_on,omitempty" bson:"created_on,omitempty"`
	UpdatedOn *time.Time `json:"updated_on,omitempty" bson:"updated_on,omitempty"`
}

type Repository struct {
	Model       `json:",inline"`
	ID          int64  `json:"id" bson:"_id"`
	Owner       string `json:"owner" bson:"owner"`
	Name        string `json:"name" bson:"name"`
	URL         string `json:"url" bson:"url"`
	Description string `json:"description" bson:"description"`
}

type BountyState int

const (
	BountyStateOpen BountyState = iota
	BountyStateReleased
	BountyStateTransferred
)

type Bounty struct {
	Model           `json:",inline"`
	ID              int64       `json:"id" bson:"_id"`
	IssueNumber     int         `json:"issue_number" bson:"issue_number"`
	RepositoryID    int64       `json:"repository_id" bson:"repository_id"`
	ReceiverID      int64       `json:"receiver_id" bson:"receiver_id"`
	Seed            string      `json:"-" bson:"seed"`
	PoolAddress     string      `json:"pool_address" bson:"pool_address"`
	ReceiverAddress string      `json:"receiver_address" bson:"receiver_address"`
	BundleHash      string      `json:"bundle_hash" bson:"bundle_hash"`
	Balance         uint64      `json:"balance" bson:"balance"`
	URL             string      `json:"url" bson:"url"`
	Title           string      `json:"title" bson:"title"`
	Body            string      `json:"body" bson:"body"`
	State           BountyState `json:"state" bson:"state"`
}

// Used to circumvent duplicated _id fields
type DeletedModel struct {
	Object interface{} `json:"object" bson:"object"`
}
