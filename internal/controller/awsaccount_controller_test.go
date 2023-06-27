package controller

import (
	"context"
	"errors"
	"time"

	kuadrav1 "github.com/Kuadrant/kuadra/api/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8Types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("AwsAccount controller", func() {

	const (
		AwsAccountName      = "awsaccount-test"
		AwsAccountNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	ctx := context.Background()

	awsController := &kuadrav1.AwsAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AwsAccount",
			APIVersion: "kuadra.kuadrant.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      AwsAccountName,
			Namespace: AwsAccountNamespace,
		},
		Spec: kuadrav1.AwsAccountSpec{
			UserName: "ib-dns",
			Groups: []string{
				"dns-management",
				"test-group",
			},
			Zones: []string{
				"ib.kuadra.io",
			},
		},
	}

	awsAccountLookupKey := k8Types.NamespacedName{Name: AwsAccountName, Namespace: AwsAccountNamespace}
	createdAwsAccount := &kuadrav1.AwsAccount{
		Spec: kuadrav1.AwsAccountSpec{
			UserName: "ib-dns",
		},
		Status: kuadrav1.AwsAccountStatus{},
	}

	Context("When checking AwsController Status", func() {
		It("Should run AwsAccount reconcile", func() {
			By("By checking reconcile")
			req := reconcile.Request{
				NamespacedName: awsAccountLookupKey,
			}

			client := fake.NewClientBuilder().Build()

			mockIam := mockIamWrapper{
				Users:        []types.User{},
				LoginProfile: map[string]types.LoginProfile{},
				AccessKey:    map[string]types.AccessKey{},
			}

			client.Create(ctx, awsController)

			r := &AwsAccountReconciler{
				Client:     client,
				Scheme:     scheme.Scheme,
				IamWrapper: &mockIam,
			}

			_, err := r.Reconcile(ctx, req)
			if err != nil {
				Expect(err).Should(BeNil())
			}

			By("By checking if AwsAccount status is correct")
			Eventually(func() kuadrav1.AccountStatus {
				err := client.Get(ctx, awsAccountLookupKey, createdAwsAccount)
				if err != nil {
					return ""
				}
				return createdAwsAccount.Status.Account
			}, timeout, interval).Should(Equal(kuadrav1.Created))

			By("By checking created user")
			Expect(mockIam.Users).Should(Equal([]types.User{
				{
					UserName: &awsController.Spec.UserName,
				},
			}))
		})

	})
})

type mockIamWrapper struct {
	Users        []types.User
	LoginProfile map[string]types.LoginProfile
	AccessKey    map[string]types.AccessKey
}

func (c mockIamWrapper) GetUser(userName string) (*types.User, error) {
	for _, user := range c.Users {
		if *user.UserName == userName {
			return &user, nil
		}
	}
	return nil, errors.New("User not found")
}

func (c *mockIamWrapper) CreateUser(userName string) (*types.User, error) {
	user := types.User{
		UserName: &userName,
	}
	c.Users = append(c.Users, user)
	return &user, nil
}

// Not yet used in AwsController
func (c mockIamWrapper) ListUsers(maxUsers int32) ([]types.User, error) {
	return c.Users, nil
}

func (c mockIamWrapper) CreateLoginProfile(password string, userName string, passwordResetRequired bool) (types.LoginProfile, error) {
	if _, exists := c.LoginProfile[userName]; exists {
		return types.LoginProfile{}, errors.New("username already exists")
	}
	loginProfile := types.LoginProfile{
		UserName:              &userName,
		PasswordResetRequired: passwordResetRequired,
	}
	c.LoginProfile[userName] = loginProfile
	return c.LoginProfile[userName], nil
}

func (c mockIamWrapper) CreateAccessKeyPair(userName string) (*types.AccessKey, error) {
	if _, exists := c.AccessKey[userName]; exists {
		return nil, errors.New("access key already exists for the user")
	}

	accessKey := types.AccessKey{
		AccessKeyId:     aws.String("AccessKeyId"),
		SecretAccessKey: aws.String("SecretAccessKey"),
	}
	c.AccessKey[userName] = accessKey

	return &accessKey, nil
}

// Not yet used in AwsController
func (c mockIamWrapper) AddUserToGroup(groupName string, userName string) (middleware.Metadata, error) {
	return middleware.Metadata{}, nil
}
