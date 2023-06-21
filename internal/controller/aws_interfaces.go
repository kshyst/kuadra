package controller

import (
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
)

type IamWrapper interface {
	GetUser(userName string) (*types.User, error)
	CreateUser(userName string) (*types.User, error)
	ListUsers(maxUsers int32) ([]types.User, error)
	CreateLoginProfile(password string, userName string, passwordResetRequired bool) (types.LoginProfile, error)
	CreateAccessKeyPair(userName string) (*types.AccessKey, error)
	AddUserToGroup(groupName string, userName string) (middleware.Metadata, error)
}
