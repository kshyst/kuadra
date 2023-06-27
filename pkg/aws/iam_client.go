package aws

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type iamWrapper struct {
	IamClient *iam.Client
}

func NewIamWrapper() (*iamWrapper, error) {
	// TODO: take config/credentials in this constructor
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		return nil, err
	}

	iamWrapper := iamWrapper{
		IamClient: iam.NewFromConfig(sdkConfig),
	}
	return &iamWrapper, nil
}

func (wrapper iamWrapper) GetUser(userName string) (*types.User, error) {
	var user *types.User
	result, err := wrapper.IamClient.GetUser(context.TODO(), &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				log.Printf("User %v does not exist.\n", userName)
				err = nil
			default:
				log.Printf("Couldn't get user %v. Here's why: %v\n", userName, err)
			}
		}
	} else {
		user = result.User
	}
	return user, err
}

func (wrapper iamWrapper) IsExistingUser(userName string) (bool, error) {
	_, err := wrapper.IamClient.GetUser(context.TODO(), &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				return false, nil
			default:
				return false, err
			}
		}
	}
	return true, nil
}

func (wrapper iamWrapper) HasLoginProfile(userName string) (bool, error) {
	_, err := wrapper.IamClient.GetLoginProfile(context.TODO(), &iam.GetLoginProfileInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NoSuchEntityException:
				return false, nil
			default:
				return false, err
			}
		}
	}
	return true, nil
}

func (wrapper iamWrapper) HasAccessKey(userName string) (bool, error) {
	result, err := wrapper.IamClient.ListAccessKeys(context.TODO(), &iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return false, err
	}
	return len(result.AccessKeyMetadata) > 0, nil
}

func (wrapper iamWrapper) ListGroupsForUser(userName string) ([]types.Group, error) {
	result, err := wrapper.IamClient.ListGroupsForUser(context.TODO(), &iam.ListGroupsForUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}
	return result.Groups, nil
}

func (wrapper iamWrapper) CreateUser(userName string) (*types.User, error) {
	var user *types.User
	result, err := wrapper.IamClient.CreateUser(context.TODO(), &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		log.Printf("Couldn't create user %v. Here's why: %v\n", userName, err)
	} else {
		user = result.User
	}
	return user, err
}

func (wrapper iamWrapper) CreateUserIfNotExists(userName string) error {
	_, err := wrapper.IamClient.CreateUser(context.TODO(), &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.EntityAlreadyExistsException:
				return nil
			default:
				return err
			}
		}
	}
	return nil
}

func (wrapper iamWrapper) ListUsers(maxUsers int32) ([]types.User, error) {
	var users []types.User
	result, err := wrapper.IamClient.ListUsers(context.TODO(), &iam.ListUsersInput{
		MaxItems: aws.Int32(maxUsers),
	})
	if err != nil {
		log.Printf("Couldn't list users. Here's why: %v\n", err)
	} else {
		users = result.Users
	}
	return users, err
}

func (wrapper iamWrapper) CreateLoginProfile(password string, userName string, passwordResetRequired bool) (types.LoginProfile, error) {
	result, err := wrapper.IamClient.CreateLoginProfile(context.TODO(), &iam.CreateLoginProfileInput{
		Password:              &password,
		UserName:              &userName,
		PasswordResetRequired: passwordResetRequired,
	})
	var loginProfile types.LoginProfile
	if err != nil {
		log.Printf("Couldn't create login profile. Here's why: %v\n", err)
	} else {
		loginProfile = *result.LoginProfile
	}
	return loginProfile, err
}

func (wrapper iamWrapper) CreateLoginProfileIfNotExists(password string, userName string, passwordResetRequired bool) error {
	_, err := wrapper.IamClient.CreateLoginProfile(context.TODO(), &iam.CreateLoginProfileInput{
		Password:              &password,
		UserName:              &userName,
		PasswordResetRequired: passwordResetRequired,
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.EntityAlreadyExistsException:
				return nil
			default:
				return err
			}
		}
	}
	return nil
}

func (wrapper iamWrapper) CreateAccessKeyPair(userName string) (*types.AccessKey, error) {
	var key *types.AccessKey
	result, err := wrapper.IamClient.CreateAccessKey(context.TODO(), &iam.CreateAccessKeyInput{
		UserName: aws.String(userName)})
	if err != nil {
		log.Printf("Couldn't create access key pair for user %v. Here's why: %v\n", userName, err)
	} else {
		key = result.AccessKey
	}
	return key, err
}

func (wrapper iamWrapper) AddUserToGroup(groupName string, userName string) (middleware.Metadata, error) {
	var metadata middleware.Metadata
	result, err := wrapper.IamClient.AddUserToGroup(context.TODO(), &iam.AddUserToGroupInput{
		GroupName: aws.String(groupName),
		UserName:  aws.String(userName)})
	if err != nil {
		log.Printf("Couldn't add user %v to group. Here's why: %v\n", userName, err)
	} else {
		metadata = result.ResultMetadata
	}
	return metadata, err
}

func (wrapper iamWrapper) RemoveUserFromGroup(groupName string, userName string) (middleware.Metadata, error) {
	var metadata middleware.Metadata
	result, err := wrapper.IamClient.RemoveUserFromGroup(context.TODO(), &iam.RemoveUserFromGroupInput{
		GroupName: aws.String(groupName),
		UserName:  aws.String(userName)})
	if err != nil {
		log.Printf("Couldn't remove user %v from group %v. Here's why: %v\n", userName, groupName, err)
	} else {
		metadata = result.ResultMetadata
	}
	return metadata, err
}
