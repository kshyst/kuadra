package controller

import (
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
)

type IamWrapper interface {
	GetUser(userName string) (*types.User, error)
	IsExistingUser(userName string) (bool, error)
	HasLoginProfile(userName string) (bool, error)
	HasAccessKey(userName string) (bool, error)
	ListGroupsForUser(userName string) ([]types.Group, error)
	CreateUser(userName string) (*types.User, error)
	CreateUserIfNotExists(userName string) error
	ListUsers(maxUsers int32) ([]types.User, error)
	CreateLoginProfile(password string, userName string, passwordResetRequired bool) (types.LoginProfile, error)
	CreateLoginProfileIfNotExists(password string, userName string, passwordResetRequired bool) error
	CreateAccessKeyPair(userName string) (*types.AccessKey, error)
	AddUserToGroup(groupName string, userName string) (middleware.Metadata, error)
	RemoveUserFromGroup(groupName string, userName string) (middleware.Metadata, error)
}
