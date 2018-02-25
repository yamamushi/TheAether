[![Travis](https://img.shields.io/travis/yamamushi/TheAether.svg?branch=master)](https://travis-ci.org/yamamushi/TheAether/) [![GitHub issues](https://img.shields.io/github/issues/yamamushi/TheAether.svg)](https://GitHub.com/yamamushi/TheAether/issues/) [![Go Report Card](https://goreportcard.com/badge/github.com/yamamushi/TheAether)](https://goreportcard.com/report/github.com/yamamushi/TheAether?style=plastic) [![Website shields.io](https://img.shields.io/website-up-down-green-red/http/TheAetherGame.com.svg)](http://shields.io/) 
 
 [![Build history](https://buildstats.info/travisci/chart/yamamushi/TheAether)](https://ci.appveyor.com/project/yamamushi/TheAether/history)




The Aether
=========

[![The Aether Central Hub](https://img.shields.io/badge/Play%20The%20Aether-%23welcome-blue.svg)](https://discord.gg/Ju28dpe)

The Aether is a roleplaying game and open source MUD framework developed for Discord.

Will you become a traveled adventurer or a rich king? Perhaps a ship merchantman or a shopkeeper? Whatever you choose to become, The Aether welcomes you on your journey!




## Table of Contents

   * [About](#About)
   * [Features](#features)
   * [Commands](#commands)
     * [Role Management](#role-management-commands)
     * [Room Management](#room-management-commands)
     * [Cluster Management](#cluster-management-commands)
     * [Notify Command](#notify-command)
     * [Events Command](#events-command)
   * [Development](#development)
   * [Discord](#discord)


## About 

The Aether is, at its core, a MUD that runs on top of Discord. However, what separates The Aether from other discord roleplaying games (such as Discord RPG, which is arguably a great game on its own) is that by playing it you are a true participant in the world.

While other discord bots control traveling in a 2-dimensional way (you can play without ever leaving a channel, and other participants can be in the same chanel as you), The Aether controls traveling in a 3-dimensional way through the world by managing the roles and permissions that define Discord. That is to say, when you travel "north" from a room, for example, roles are assigned and revoked from your account with varying permissions that emulate the feeling of actually moving to a different location.


 

## Features

**Completed** 

- [X] Multiple discord linking (creating a web of discords for nearly unlimited world size)
- [X] Room Creation and linking
- [X] Traveling between rooms 



**Planned**

- [ ] Character Creation
- [ ] Item creation with different item types
- [ ] Traveling Creatures
- [ ] NPC Management
- [ ] Currency System 



## Commands

### Role Management Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| perms addrole |  | |
| perms removerole |  |  | 
| perms createrole |  |  | 
| perms deleterole |  |  | 
| perms viewrole |  |  | 
| perms syncserverroles |  |  | 
| perms syncrolesdb | | | 
| perms translaterole |  |  |


### Room Management Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| room add | | |
| room remove | | |
| room roles | | |
| room travelrole |  |  |
| room travelroleclear |  |  |
| room view |  |  |
| room linkrole |  |  |
| room unlinkrole |  |  |
| room setupserver |  |  |
| room description |  |  |
| room guildinvite |  |  |
| room linkdirection |  |  |


### Cluster Management Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| guilds sync cluster | sync and repair all guilds in the cluster | |
| guilds sync guild | sync and repair specific guild | |
| guilds info | display information about a guild | |
| guilds cluster | display cluster stats | |   


### Notify Command

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| enable |  | |
| enable for |  |  |
| disable  |  |  |
| disable for |  |  |
| add |  |  |
| remove |  |  |
| list |  |  |
| view |  |  |
| channel |  |  |
| messages |  |  |
| flush |  |  |
| linked |  |  |



### Events Command

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| add |  |  |
| remove |  |  |
| list |  |  |
| info |  |  |
| enable | | |
| disable | | |
| listenabled | | |


## Development

**Development Branch Status**

[![Travis](https://img.shields.io/travis/yamamushi/TheAether.svg?branch=develop)](https://travis-ci.org/yamamushi/TheAether/)


## Discord

Join us on Discord 

![Discord](https://img.shields.io/discord/414332617454780416.svg) 


[![Open Source Love svg1](https://badges.frapsoft.com/os/v1/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)