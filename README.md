# Kuadra

## What is it?

A kubernetes controller for managing users and access permissions in various services.
It will watch ConfigMaps or custom resources that contain user configuration.
The controller's job is to reconcile that config by making API calls to various services (such as AWS) to ensure a team (i.e. a set of users) has accounts and access set up correctly in those services.
The config will be declarative, so the controller will also take care of updating or deleting things in those various services as well.

## Features

Initially it will ensure an AWS user account exists for each user in a specific AWS org, they have a hosted zone with permissions to create DNS records, and can generate access keys.

## Kuadra name

It’s a combination of Kuadrant and Hydra.
Hydra being the mythical serpentine monster with many heads. (Kuadra will have integrations into many things)
Hydra is also known for for its regenerative abilities (Kuadra will have a reconcile loop for ‘self healing’)

## Running the Operator

Before working on the project you should have a good idea of the technologies used such as [Go](https://go.dev/learn/), [Kubernetes](https://kubernetes.io/docs/setup/), and building operators for Kubernetes clusters. This project uses [kubebuilder](https://book.kubebuilder.io/getting-started) to build the operator, so take a look at the [quick start](https://book.kubebuilder.io/quick-start) to get accustomed to it if you haven't already.

There are two ways we recommend to run the operator for testing. The first way is running locally on a [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster, and the second way is running a containerised version of the operator locally with [Docker](https://docs.docker.com/guides/get-started/). Both ways are described in the Makefile.

In order to run the cluster, a few pre-requisites are required. You should have kind, kubectl, Docker, and Go installed prior to following the steps.

Before following the steps below, please make sure you have aws-cli [set up and configured](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html#getting-started-quickstart-new-command) with your access key. Also ensure you have the a policy attached to your AWS IAM user that contains at least the following actions.

```json
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "VisualEditor0",
			"Effect": "Allow",
			"Action": [
				"iam:CreateLoginProfile",
				"iam:ListGroupsForUser",
				"iam:GetUser",
				"iam:CreateUser",
				"iam:GetLoginProfile",
				"iam:ListAccessKeys",
				"iam:CreateAccessKey",
				"iam:AddUserToGroup",
				"iam:RemoveUserFromGroup",
				"iam:DeleteLoginProfile",
				"iam:DeleteAccessKey",
				"iam:DeleteUser"
			],
			"Resource": "*"
		}
	]
}
```

Once pre-requisites are installed, you can run the following commands to get the operator up and running.

### Running outside a kind cluster (kind required)
```bash
# 1. Create the cluster using Kind (Kubernetes in Docker)
kind create cluster
# 2. Install CRD's
make install
# 3. Disable webhooks (throws error if webhook is enabled)
export ENABLE_WEBHOOKS=false
# 4. Run operator locally
make run
# 5. Add sample config to your cluster in the default namespace.
kubectl apply -k config/samples
```

### Running locally in a kind cluster

Before following the below instructions, please ensure you have docker-cli installed and configured with your [quay.io account](https://docs.quay.io/solution/getting-started.html), as you will need to push a built image to your own namespace/account. By default, quay.io will set the visibility of your repository to private. In order for your cluster pods to pull the image, you will need to set the visibility of your repository to public after pushing your image. You can do this in your repository settings.

Also, before following the steps below ensure you have created a `aws-credentials.env` file in the root directory of your repo clone with your access key credentials. Refer to the above reference to AWS Credentials for information on how to set up an access key. The file should follow the format:

```
AWS_ACCESS_KEY_ID=<your aws access key id>
AWS_SECRET_ACCESS_KEY=<your aws secret access key>
```

```bash
# 1. Create the cluster using Kind (Kubernetes in Docker)
kind create cluster
# 2. Install CRD's
make install
# 3. Set the IMG variable to where you would like to push your image to, then build and push the image, then deploy.
IMG=quay.io/<namespace>/kuadra:v1 make docker-build docker-push deploy
# 4. Add sample config to your cluster in the kuadra-system namespace.
kubectl -n kuadra-system apply -k config/samples
```

