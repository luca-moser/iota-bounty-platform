package models

import (
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Model struct {
	CreatedOn time.Time  `json:"created_on,omitempty" bson:"created_on,omitempty"`
	UpdatedOn *time.Time `json:"updated_on,omitempty" bson:"updated_on,omitempty"`
}

type User struct {
	Model             `json:",inline"`
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	Username          string             `json:"username" bson:"username"`
	Email             string             `json:"email" bson:"email"`
	Admin             bool               `json:"admin" bson:"admin"`
	Password          string             `json:"-" bson:"password"`
	PasswordSalt      string             `json:"-" bson:"password_salt"`
	Confirmed         bool               `json:"confirmed" bson:"confirmed"`
	ConfirmationCode  string             `json:"-" bson:"confirmation_code"`
	LastAccess        time.Time          `json:"last_access" bson:"last_access"`
	LastLogin         time.Time          `json:"last_login" bson:"last_login"`
	Deactivated       bool               `json:"-" bson:"deactivated"`
	PasswordResetCode string             `json:"-" bson:"password_reset_code"`
}

type UserJWTClaims struct {
	jwt.StandardClaims
	UserID    primitive.ObjectID `json:"user_id"`
	Username  string             `json:"username"`
	Auth      bool               `json:"auth"`
	Confirmed bool               `json:"confirmed"`
	Admin     bool               `json:"admin"`
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
