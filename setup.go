package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

type SetupProcess struct {

	db       *DBHandler
	conf     *Config
	user     *UserHandler

}


func (h *SetupProcess) Init(s *discordgo.Session, channelID string) error {

	h.SetupOwnerPermissions(s,channelID)

	return nil
}


func (h *SetupProcess) SetupOwnerPermissions(s *discordgo.Session, channelID string) (err error) {
	fmt.Println("Verifying Guild Owner")
	ownerID, err := getGuildOwnerID(s, channelID)
	if err != nil {
		return err
	}

	db := h.db.rawdb.From("Users")

	var user User
	err = db.One("ID", ownerID, &user)
	if err != nil {

		fmt.Println("Verifying owner in DB: " + ownerID)

		guildID, err := getGuildID(s, channelID)
		if err != nil {
			return err
		}

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

