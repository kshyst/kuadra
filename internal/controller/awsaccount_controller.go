/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/sethvargo/go-password/password"

	kuadrav1 "github.com/Kuadrant/kuadra/api/v1"
)

func contains[T comparable](slice []T, val T) bool {
	for _, element := range slice {
		if element == val {
			return true
		}
	}
	return false
}

func indexOf[T comparable](slice []T, element T) int {
	for i, val := range slice {
		if val == element {
			return i
		}
	}
	return -1
}

func remove[T any](slice []T, i int) []T {
	return append(slice[:i], slice[i+1:]...)
}

func getLeftDifference[T comparable](left []T, right []T) (leftDifference []T) {
	for _, val := range left {
		if !contains[T](right, val) {
			leftDifference = append(leftDifference, val)
		}
	}
	return leftDifference
}

// AwsAccountReconciler reconciles a AwsAccount object
type AwsAccountReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	IamWrapper IamWrapper
}

//+kubebuilder:rbac:groups=kuadra.kuadrant.io,resources=awsaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuadra.kuadrant.io,resources=awsaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuadra.kuadrant.io,resources=awsaccounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AwsAccount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *AwsAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var previous kuadrav1.AwsAccount
	if err := r.Get(ctx, req.NamespacedName, &previous); err != nil {
		log.Error(err, "unable to fetch AwsAccount")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	awsAccount := previous.DeepCopy()

	refreshedStatus, err := r.getRefreshedStatus(*awsAccount)
	if err != nil {
		log.Error(err, "unable to get refreshed status")
		return ctrl.Result{}, err
	}
	awsAccount.Status = *refreshedStatus

	if !awsAccount.Status.UserCreated {
		if err := r.IamWrapper.CreateUserIfNotExists(awsAccount.Spec.UserName); err != nil {
			log.Error(err, "unable to create IAM user")
			return ctrl.Result{}, err
		}
		log.V(1).Info("created user", "userName", awsAccount.Spec.UserName)
		awsAccount.Status.UserCreated = true
	}

	if !awsAccount.Status.LoginProfileCreated {
		pass, err := password.Generate(20, 3, 3, false, true)
		if err != nil {
			log.Error(err, "unable to generate password")
			return ctrl.Result{}, err
		}
		if err := r.IamWrapper.CreateLoginProfileIfNotExists(pass, awsAccount.Spec.UserName, true); err != nil {
			log.Error(err, "unable to create login profile")
			return ctrl.Result{}, err
		}
		log.V(1).Info("created login profile")
		awsAccount.Status.LoginProfileCreated = true

		// TODO: Save this to a secret in user's namespace
		log.V(1).Info("Temporary password", "password", pass)
	}

	if !awsAccount.Status.AccessKeyCreated {
		accessKey, err := r.IamWrapper.CreateAccessKeyPair(awsAccount.Spec.UserName)
		if err != nil {
			log.Error(err, "unable to create access key")
			return ctrl.Result{}, err
		}
		log.V(1).Info("created access key", "accessKeyId", accessKey.AccessKeyId)
		awsAccount.Status.AccessKeyCreated = true

		// TODO: Save these to a secret in user's namespace
		log.V(1).Info("Credentials", "accessKeyId:", accessKey.AccessKeyId, "accessKeySecret", accessKey.SecretAccessKey)
	}

	groupsToAddUserTo := getLeftDifference[string](awsAccount.Spec.Groups, awsAccount.Status.UserGroups)
	for _, group := range groupsToAddUserTo {
		if _, err := r.IamWrapper.AddUserToGroup(group, awsAccount.Spec.UserName); err != nil {
			log.Error(err, "unable to add user to group", "groupName", group)
			return ctrl.Result{}, err
		}
		log.V(1).Info("Added user to group", "group name:", group)
		awsAccount.Status.UserGroups = append(awsAccount.Status.UserGroups, group)
	}

	groupsToRemoveUserFrom := getLeftDifference[string](awsAccount.Status.UserGroups, awsAccount.Spec.Groups)
	for _, group := range groupsToRemoveUserFrom {
		if _, err := r.IamWrapper.RemoveUserFromGroup(group, awsAccount.Spec.UserName); err != nil {
			log.Error(err, "unable to remove user from group", "groupName", group)
			return ctrl.Result{}, err
		}
		log.V(1).Info("removed user from group", "groupName", group)
		index := indexOf[string](awsAccount.Status.UserGroups, group)
		awsAccount.Status.UserGroups = remove[string](awsAccount.Status.UserGroups, index)
	}

	if !reflect.DeepEqual(previous, *awsAccount) {
		if err := r.Status().Update(ctx, awsAccount); err != nil {
			log.Error(err, "unable to update awsAccount status")
			return ctrl.Result{RequeueAfter: time.Second * 3}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AwsAccountReconciler) getRefreshedStatus(awsAccount kuadrav1.AwsAccount) (*kuadrav1.AwsAccountStatus, error) {
	var status kuadrav1.AwsAccountStatus
	userExists, err := r.IamWrapper.IsExistingUser(awsAccount.Spec.UserName)
	if err != nil {
		return nil, err
	}
	if !userExists {
		// Return struct with zero values
		return &status, nil
	}
	status.UserCreated = true

	loginProfileExists, err := r.IamWrapper.HasLoginProfile(awsAccount.Spec.UserName)
	if err != nil {
		return nil, err
	}
	status.LoginProfileCreated = loginProfileExists

	accessKeyExists, err := r.IamWrapper.HasAccessKey(awsAccount.Spec.UserName)
	if err != nil {
		return nil, err
	}
	status.AccessKeyCreated = accessKeyExists

	groups, err := r.IamWrapper.ListGroupsForUser(awsAccount.Spec.UserName)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		status.UserGroups = append(status.UserGroups, *group.GroupName)
	}
	return &status, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuadrav1.AwsAccount{}).
		Complete(r)
}
