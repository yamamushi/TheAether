package main

import (

	"sync"
	"errors"
	"strings"
	"github.com/bwmarrin/discordgo"
)

type GuildsManager struct {

	db          *DBHandler
	querylocker sync.RWMutex

}


type GuildRecord struct {

	ID 			string `storm:"id"` // primary key
	Name 		string

	Roles		[]string
	Members		[]string

}



func (h *GuildsManager) SaveGuildToDB(guild GuildRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Guilds")
	err = db.Save(&guild)
	return err
}

func (h *GuildsManager) RemoveGuildFromDB(guild GuildRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Guilds")
	err = db.DeleteStruct(&guild)
	return err
}

func (h *GuildsManager) RemoveGuildByID(guildID string) (err error) {

	guild, err := h.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	err = h.RemoveGuildFromDB(guild)
	if err != nil {
		return err
	}

	return nil
}

func (h *GuildsManager) GetGuildByID(guildID string) (guild GuildRecord, err error) {

	guilds, err := h.GetAllGuilds()
	if err != nil{
		return guild, err
	}

	for _, i := range guilds {
		if i.ID == guildID{
			return i, nil
		}
	}

	return guild, errors.New("No guild record found")
}

func (h *GuildsManager) GetGuildByName(guildname string, guildID string) (guild GuildRecord, err error) {

	guilds, err := h.GetAllGuilds()
	if err != nil{
		return guild, err
	}

	for _, i := range guilds {
		if i.Name == guildname && i.ID == guildID{
			return i, nil
		}
	}

	return guild, errors.New("No guild record found")
}


// GetAllRooms function
func (h *GuildsManager) GetAllGuilds() (guildlist []GuildRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Guilds")
	err = db.All(&guildlist)
	if err != nil {
		return guildlist, err
	}

	return guildlist, nil
}


func (h *GuildsManager) AddRoleToGuild(guildID string, roleID string) (err error) {

	guild, err := h.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	for _, role := range guild.Roles {
		if role == roleID {
			return nil
		}
	}

	guild.Roles = append(guild.Roles, roleID)

	err = h.SaveGuildToDB(guild)
	if err != nil {
		return err
	}

	return nil
}


func (h *GuildsManager) RemoveRoleFromGuild(guildID string, roleID string) (err error) {
	guild, err := h.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	guild.Roles = RemoveStringFromSlice(guild.Roles, roleID)

	err = h.SaveGuildToDB(guild)
	if err != nil {
		return err
	}

	return nil
}



func (h *GuildsManager) AddUserToGuild(guildID string, userID string) (err error) {

	guild, err := h.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	for _, user := range guild.Members {
		if user == userID {
			return nil
		}
	}

	guild.Members = append(guild.Members, userID)

	err = h.SaveGuildToDB(guild)
	if err != nil {
		return err
	}

	return nil
}


func (h *GuildsManager) RemoveUserFromGuild(guildID string, userID string) (err error) {
	guild, err := h.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	guild.Members = RemoveStringFromSlice(guild.Members, userID)

	err = h.SaveGuildToDB(guild)
	if err != nil {
		return err
	}

	return nil
}


func (h *GuildsManager) RegisterGuild(guildID string, s *discordgo.Session) (err error) {

	guildRecord, err := h.GetGuildByID(guildID)
	if err != nil {
		if strings.Contains(err.Error(), "No guild record found"){

			discordguild, err := s.Guild(guildID)
			if err != nil {
				return err
			}

			guildRecord = GuildRecord{ID: guildID, Name: discordguild.Name }
			err = h.SaveGuildToDB(guildRecord)
			if err != nil {
				return err
			}

			return nil

		} else {
			return err
		}
	} else {
		return nil
	}
}