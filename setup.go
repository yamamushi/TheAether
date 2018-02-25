package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

// SetupProcess function
type SetupProcess struct {
	db     *DBHandler
	conf   *Config
	user   *UserHandler
	rooms  *RoomsHandler
	guilds *GuildsManager
}

// Init function
func (h *SetupProcess) Init(s *discordgo.Session, channelID string) (err error) {

	err = h.SetupOwnerPermissions(s, channelID)
	if err != nil {
		return err
	}

	fmt.Println("\n|| Running Rooms Setup ||\n ")

	err = h.rooms.InitRooms(s, channelID)
	if err != nil {
		fmt.Println("Init Rooms: " + err.Error())
	}

	return nil
}

// SetupOwnerPermissions function
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
		fmt.Println("\n\n!!! The bot must first be setup on the main cluster server to be configured properly")
		fmt.Println("!!! The owner ID of the main server must also be configured properly.")
		return errors.New("Could not complete setup")
	}

	fmt.Println("Getting Guild Record")
	guildRecord, err := h.guilds.GetGuildByID(guildID)
	if err != nil {
		if strings.Contains(err.Error(), "No guild record found") {
			fmt.Println("Registering Guild in Database")

			discordguild, err := s.Guild(guildID)
			if err != nil {
				return err
			}

			guildRecord = GuildRecord{ID: guildID, Name: discordguild.Name}
			err = h.guilds.SaveGuildToDB(guildRecord)
			if err != nil {
				return err
			}

			fmt.Println("Guild Registered: " + guildRecord.ID + " - " + guildRecord.Name)

		} else {
			fmt.Println("Error retrieving guild: " + err.Error())
			return err
		}
	} else {
		fmt.Println("Guild Registered: " + guildRecord.ID + " - " + guildRecord.Name)
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
			fmt.Println("Error inserting usermanager into Database!")
			return err
		}
		fmt.Println("Owner ID: " + ownerID)

	}
	return nil
}
