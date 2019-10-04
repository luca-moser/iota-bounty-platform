# IOTA Bounty Platform

A platform for linking issues on GitHub to bounties which are payed out using IOTA.

> This product is in an alpha stage

Features:
* Use a GitHub account to post messages on linked issues with status updates
* Let people pool money on a particular issue to increase the incentive to solve the problem/add the feature.
* Feeless transfer of the bounty to the recipient through IOTA
* Single-page-application to manage linked repositories and bounties
* Written using modern technologies such as TypeScript, React, MobX and Go

![Bounties](https://i.imgur.com/kyl8MFW.png)

### Installation

Prerequisites:
* Ubuntu +18.04
* Docker
* Docker Compose

#### Create a bot user
1. Create a new user account on GitHub which is going to be used for the bot handling bounties/messaging from
and to the backend system
2. Generate a personal access token for the bot with permissions for repository and user related actions
3. Keep the token for later installation instructions

![permissions](https://i.imgur.com/ssVjiTy.png)

#### Setting up the docker image


A docker image is provided under `lucamoser/ibp:<version>`.

