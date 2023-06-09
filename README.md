# Kuadra

## What is it?

A kubernetes controller for managing users and access permissions in various services.
It will watch ConfgMaps or custom resources that contain user configuration.
The controller's job is to reconcile that config by making API calls to various services (such as AWS) to ensure a team (i.e. a set of users) has accounts and access set up correctly in those services.
The config will be declarative, so the controller will also take care of updating or deleting things in those various services as well.

## Features

Initially it will ensure an AWS user account exists for each user in a specific AWS org, they have a hosted zone with permissions to create DNS records, and can generate access keys.

## Kuadra name

It’s a combination of Kuadrant and Hydra.
Hydra being the mythical serpentine monster with many heads. (Kuadra will have integrations into many things)
Hydra is also known for for its regenerative abilities (Kuadra will have a reconcile loop for ‘self healing’)
