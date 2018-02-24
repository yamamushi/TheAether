package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"strconv"
	"time"
)

// UserHandler struct
type UserHandler struct {
	conf        *Config
	db          *DBHandler
	cp          string
	logchan     chan string
	usermanager *UserManager
}

// Init function
func (h *UserHandler) Init() {
	h.cp = h.conf.MainConfig.CP
	h.usermanager = new(UserManager)
	h.usermanager.db = h.db
}

// Read function
func (h *UserHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(m.Content)

	if len(message) < 1 {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}

	// We use this a bit, this is the author id formatted as a mention
	mention := m.Author.Mention()

	if message[0] == cp+"groups" {
		mentions := m.Mentions

		if len(mentions) == 0 {
			groups, err := h.GetGroups(user.ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		if len(message) == 2 {
			groups, err := h.GetGroups(mentions[0].ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "groups" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}
	}

	if message[0] == cp+"roles" {
		mentions := m.Mentions

		if len(mentions) == 0 {
			roles, err := h.GetRoles(user.ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving roles: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatRoles(roles))
			return
		}

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		if len(message) == 2 {
			roles, err := h.GetRoles(mentions[0].ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving roles: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "roles" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatRoles(roles))
			return
		}
	}

	if message[0] == cp+"repairuser" {
		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		mentions := m.Mentions

		if len(message) == 2 {
			err := h.RepairUser(mentions[0].ID, s, m.ChannelID, "")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing usermanager: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, ":construction: User record repaired!")
			return
		}

		if len(message) == 3 {
			err := h.RepairUser(mentions[0].ID, s, m.ChannelID, message[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing usermanager: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, ":construction: User record repaired!")
			return
		}

	}

	if message[0] == cp+"debuguser"{
		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}
		mentions := m.Mentions

		if len(message) == 2 {
			err := h.DebugUser(mentions[0].ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing usermanager: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			return
		}
	}

	if message[0] == cp+"attributes"{
		/*
		if !usermanager.CheckRole("player") || !usermanager.CheckRole("Registered") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}
		*/
		if len(message) > 1 {
			if !user.CheckRole("moderator")  {

				attributes := h.GetFormattedAttributes(m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Attributes: \n" + attributes)
				return
			} else {

				mentionlist := m.Mentions
				if len(mentionlist) < 1 {
					s.ChannelMessageSend(m.ChannelID, ":exclamation: Invalid user mention!")
					return
				}

				attributes := h.GetFormattedAttributes(mentionlist[0].ID)
				s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Attributes: \n" + attributes)
				return
			}
		} else {
			attributes := h.GetFormattedAttributes(m.Author.ID)
			s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Attributes: \n" + attributes)
			return
		}
	}
	if message[0] == cp+"stats"{
		/*
		if !usermanager.CheckRole("player") || !usermanager.CheckRole("Registered") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}
		*/
		if len(message) > 1 {
			if !user.CheckRole("moderator")  {

				stats := h.GetFormattedStats(m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Stats: \n" + stats)
				return
			} else {
				mentionlist := m.Mentions
				if len(mentionlist) < 1 {
					s.ChannelMessageSend(m.ChannelID, ":exclamation: Invalid user mention!")
					return
				}
				stats := h.GetFormattedStats(mentionlist[0].ID)
				s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Stats: \n" + stats)
				return
			}
		} else {
			stats := h.GetFormattedStats(m.Author.ID)
			s.ChannelMessageSend(m.ChannelID, ":large_blue_diamond: Stats: \n" + stats)
			return
		}
	}

	return
}


func (h *UserHandler) AddItem(itemid string, userid string,  s *discordgo.Session, channelID string) (err error) {

	// Make sure usermanager is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	user.ItemsMap = append(user.ItemsMap, itemid)
	return nil
}


func (h *UserHandler) RemoveItem(itemid string, userid string,  s *discordgo.Session, channelID string) (err error) {

	// Make sure usermanager is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	user.ItemsMap = RemoveStringFromSlice(user.ItemsMap, itemid)
	return nil
}


func (h *UserHandler) RepairUser(userid string, s *discordgo.Session, channelID string, guildID string) (err error) {

	// Make sure usermanager is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	if guildID == "" {
		guildID, err = getGuildID(s, channelID)
		if err != nil {
			return err
		}
	}
	user.GuildID = guildID

	// These are ID's now!
	for _, roleID := range user.RoleIDs {
		// Ignore errors if role doesn't exist on the guild
		time.Sleep(time.Duration(time.Second*1))
		_ = s.GuildMemberRoleAdd(guildID, user.ID, roleID)
	}

	db := h.db.rawdb.From("Users")

	err = db.Update(&user)
	if err != nil {
		fmt.Println("Error updating usermanager record into database!")
		return
	}

	return nil
}


func (h *UserHandler) DebugUser(userid string, s *discordgo.Session, channelID string) (err error) {

	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	discordUser, err := s.User(userid)
	if err != nil{
		return err
	}


	userRecord := "```\n"
	userRecord = userRecord + "Username: " + discordUser.Username + "\n"
	userRecord = userRecord + "UserID: " + user.ID + "\n"
	userRecord = userRecord + "GuildID: " + user.GuildID + "\n"
	userRecord = userRecord + "ChannelID: " + user.RoomID + "\n"
	userRecord = userRecord + "Roles: " + "\n"
	userRecord = userRecord + "```\n"

	s.ChannelMessageSend(channelID, "User Record: " + userRecord)

	return nil
}


// GetUser function
func (h *UserHandler) GetUser(userid string, s *discordgo.Session, channelID string) (user User, err error) {

	// Make sure usermanager is in the database before we pull it out!
	h.CheckUser(userid, s, channelID)

	db := h.db.rawdb.From("Users")
	err = db.One("ID", userid, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// CheckUser function
func (h *UserHandler) CheckUser(ID string, s *discordgo.Session, channelID string) {

	db := h.db.rawdb.From("Users")

	var u User
	err := db.One("ID", ID, &u)
	if err != nil {
		//fmt.Println("Adding new usermanager to DB: " + ID)

		guildID, err := getGuildID(s, channelID)
		if err != nil {
			fmt.Println("Error retrieving guildID for usermanager: " + err.Error())
			return
		}

		user := User{ID: ID, GuildID: guildID}
		user.Init()

		err = db.Save(&user)
		if err != nil {
			fmt.Println("Error inserting usermanager into Database!")
			return
		}

		return
	}
}


func (h *UserHandler) GetRoles(ID string, s *discordgo.Session, channelID string) (roles []string, err error) {

	h.CheckUser(ID, s, channelID)
	user, err := h.GetUser(ID, s, channelID)
	if err != nil {
		return roles, err
	}

	return user.RoleIDs, nil
}


// GetGroups function
func (h *UserHandler) GetGroups(ID string, s *discordgo.Session, channelID string) (groups []string, err error) {

	h.CheckUser(ID, s, channelID)
	user, err := h.GetUser(ID, s, channelID)
	if err != nil {
		return groups, err
	}
	if user.CheckRole("owner") {
		groups = append(groups, "owner")
	}
	if user.CheckRole("admin") {
		groups = append(groups, "admin")
	}
	if user.CheckRole("smoderator") {
		groups = append(groups, "smoderator")
	}
	if user.CheckRole("moderator") {
		groups = append(groups, "moderator")
	}
	if user.CheckRole("builder") {
		groups = append(groups, "builder")
	}
	if user.CheckRole("writer") {
		groups = append(groups, "writer")
	}
	if user.CheckRole("scripter") {
		groups = append(groups, "scripter")
	}
	if user.CheckRole("architect") {
		groups = append(groups, "architect")
	}
	if user.CheckRole("player") {
		groups = append(groups, "player")
	}

	return groups, nil
}

// FormatGroups function
func (h *UserHandler) FormatGroups(groups []string) (formatted string) {
	for i, group := range groups {
		if i == len(groups)-1 {
			formatted = formatted + group
		} else {
			formatted = formatted + group + ", "
		}

	}

	return formatted
}


// FormatRoles function
func (h *UserHandler) FormatRoles(roles []string) (formatted string) {

	formatted = ":satellite: ```\n"
	formatted = formatted + "Roles: "
	for i, role := range roles {
		if i == len(roles)-1 {
			formatted = formatted + role
		} else {
			formatted = formatted + role + ", "
		}
	}

	formatted = formatted + "\n```"
	return formatted
}


func (h *UserHandler) GetFormattedAttributes(userID string) (formatted string) {

	user, err := h.usermanager.GetUserByID(userID)
	if err != nil {
		return "No user record found!"
	}

	attributes := "```\n"
	attributes = attributes + "Strength: " + strconv.Itoa(user.Strength) +"\n"
	attributes = attributes + "Dexterity: " + strconv.Itoa(user.Dexterity) +"\n"
	attributes = attributes + "Constitution: " + strconv.Itoa(user.Constitution) +"\n"
	attributes = attributes + "Intelligence: " + strconv.Itoa(user.Intelligence) +"\n"
	attributes = attributes + "Wisdom: " + strconv.Itoa(user.Wisdom) +"\n"
	attributes = attributes + "Charism: " + strconv.Itoa(user.Charisma) +"\n"
	attributes = attributes + "```\n"

	return attributes
}

func (h *UserHandler) GetFormattedStats(userID string) (formatted string) {

	_, err := h.usermanager.GetUserByID(userID)
	if err != nil {
		return "No user record found!"
	}

	stats := "not implemented yet!"

	return stats
}