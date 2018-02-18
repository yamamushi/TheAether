package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"errors"
)

type SetupProcess struct {

	db       *DBHandler
	conf     *Config
	user     *UserHandler
	rooms 	 *RoomsHandler
}


func (h *SetupProcess) Init(s *discordgo.Session, channelID string) (err error) {

	err = h.SetupOwnerPermissions(s,channelID)
	if err != nil {
		return err
	}

	err = h.rooms.InitRooms(s, channelID)
	if err != nil {
		fmt.Println("Init Rooms: " + err.Error())
	}

	return nil
}


func (h *SetupProcess) SetupOwnerPermissions(s *discordgo.Session, channelID string) (err error) {
	fmt.Println("Verifying Guild Owner")
	ownerID, err := getGuildOwnerID(s, channelID)
	if err != nil {
		return err
	}

	guildID, err := getGuildID(s, channelID)
	if err != nil {
		return err
	}

	if ownerID != h.conf.MainConfig.ClusterOwnerID || guildID != h.conf.MainConfig.CentralGuildID {
		fmt.Println("\n\n!!! The bot must first be setup on the main cluster server to be configured properly\n")
		fmt.Println("!!! The owner ID of the main server must also be configured properly.")
		return errors.New("Could not complete setup")
	}

	db := h.db.rawdb.From("Users")

	var user User
	err = db.One("ID", ownerID, &user)
	if err != nil {

		fmt.Println("Verifying owner in DB: " + ownerID)

		owner := User{ID: ownerID, GuildID: guildID}
		OwnerRole(&owner)

		err = db.Save(&owner)
		if err != nil {
			fmt.Println("Error inserting user into Database!")
			return err
		}
		fmt.Println("Owner ID: " + ownerID)

	}

	return nil
}

