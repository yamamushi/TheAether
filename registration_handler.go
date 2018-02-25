package main

/*

This handler is responsible for player registration, which includes character creation.

Once a player has created a profile and is ready to enter the world, this handler will
assign the registered role and drop them into the #Crossroads room.

 */

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"errors"
	"time"
	"strconv"
)

type RegistrationHandler struct {

	callback *CallbackHandler
	conf     *Config
	db       *DBHandler
	perm     *PermissionsHandler
	registry *CommandRegistry
	dg       *discordgo.Session
	user     *UserHandler
	ch       *ChannelHandler
	rooms	 *Rooms
	guilds 	 *GuildsManager

}


func (h *RegistrationHandler) Init() {

	h.RegisterCommands()

}

func (h *RegistrationHandler) RegisterCommands() (err error) {

	h.registry.Register("register", "Register a new account", "")
	h.registry.AddGroup("register", "player")
	h.registry.AddChannel("register", h.conf.MainConfig.LobbyChannelID)

	return nil

}

// Read function
func (h *RegistrationHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	// This should register all new users, presumably we want this done here because this is the first
	// command a usermanager should have access to.
	h.user.CheckUser(m.Author.ID, s, m.ChannelID)

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		return
	}

	if strings.HasPrefix(m.Content, cp+"register") {

		if guildID != h.conf.MainConfig.CentralGuildID {
			// Ignore registration attempts in non-central guild
			return
		}

		if m.ChannelID != h.conf.MainConfig.LobbyChannelID {
			// Ignore registration attemptions in non-lobby channel
			return
		}

		if user.Registered == "" {
			//_, payload := CleanCommand(m.Content, h.conf)

			if user.CheckRole("player") {
				h.StartRegistration(s, m)
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "You are already registered! If you continue to have issues please ask an Admin for assistance.")
			return
		}
	}
	if strings.HasPrefix(m.Content, cp+"roll-attributes") {
		if user.Registered != "" {
			s.ChannelMessageSend(m.ChannelID, "You have already been registered and cannot re-roll your attributes!")
			return
		}
		h.RollAttributes(s, m)
		return
	}
	if strings.HasPrefix(m.Content, cp+"pick-race") {
		if user.Registered != "" {
			s.ChannelMessageSend(m.ChannelID, "You have already been registered and cannot change your race!")
			return
		}
		h.PickRace(s, m)
		return
	}
	if strings.HasPrefix(m.Content, cp+"raceinfo") {
		h.RaceInfo(s, m)
		return
	}
	if strings.HasPrefix(m.Content, cp+"pick-class") {
		if user.Registered != "" {
			s.ChannelMessageSend(m.ChannelID, "You have already been registered and cannot change your class!")
			return
		}
		h.PickClass(s, m)
		return
	}
	if strings.HasPrefix(m.Content, cp+"classinfo") {
		h.ClassInfo(s, m)
		return
	}

}



// StartRegistration function
func (h *RegistrationHandler) StartRegistration(s *discordgo.Session, m *discordgo.MessageCreate) {
/*
	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}
*/
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error finding usermanager: " + err.Error())
		return
	}


	welcomeMessage := ":sunrise_over_mountains: Avatar Construction Chamber ```\n"
	welcomeMessage = welcomeMessage + "You are now standing in a large chamber of light, there are no walls as far as you can tell.\n\n"
	welcomeMessage = welcomeMessage + "A faint voice begins to fill your head...\n```\n"

	userprivatechannel, err := s.UserChannelCreate(user.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error starting Registration: " + err.Error())
		return
	}

	s.ChannelMessageSend(userprivatechannel.ID, welcomeMessage)
	time.Sleep(time.Duration(time.Second*5))

	privateMessage := "*\"Hello* " + m.Author.Mention() + " *you are now standing in what is known as the avatar construction chamber.\"*\n\n"
	privateMessage = privateMessage + ""


	time.Sleep(time.Duration(time.Second*5))
	s.ChannelMessageSend(userprivatechannel.ID, privateMessage)

	privateMessage = "\"Beyond this chamber lies the beginning of your path into The Aether, a world of unlimited possibilites awaits you!\n\n"
	privateMessage = privateMessage + "Will you choose the life of a wealthy king, or a merchantman ferrying rare goods from port to port?\n\n"
	privateMessage = privateMessage + "Will you own a tavern welcoming guests and selling your own brews to everyone with a shiny coin to spare, or "
	privateMessage = privateMessage + "Will you live as a thief among the shadows sneaking through castles at night looking for rare goods treasures steal?\n\n"
	privateMessage = privateMessage + "Will you lead a cult in the shadows, or will you band together with allies to kill a god?\n\n"
	privateMessage = privateMessage + "Whatever you choose to become and wherever you choose to go, we welcome you!\""

	time.Sleep(time.Duration(time.Second*10))
	s.ChannelMessageSend(userprivatechannel.ID, privateMessage)


	privateMessage = "\"A basic avatar has now been summoned for you, however it cannot be used until you "
	privateMessage = privateMessage + "prepare it for materialization into The Aether.\n\n"
	privateMessage = privateMessage + "We will begin by assigning attributes to your avatar, followed by picking your race, class, skills, feats, "
	privateMessage = privateMessage + "and choosing a set of starter equipment.\n\n"
	privateMessage = privateMessage + "You don't need to remember all of that though, for now you can begin by typing ~roll-attributes\""

	time.Sleep(time.Duration(time.Second*10))
	s.ChannelMessageSend(userprivatechannel.ID, privateMessage)

	err = h.SetRegistrationStep("attributes", user.ID)
	if err != nil {
		s.ChannelMessageSend(userprivatechannel.ID, "Error starting Registration: " + err.Error())
		return
	}

	return
}

// SetRegistrationStep function
func (h *RegistrationHandler) SetRegistrationStep(status string, userID string) (err error){

	switch status {
		case "attributes":
			break
		case "race":
			break
		case "class":
			break
		case "complete":
			break
		case "skills":
			break
		case "feats":
			break
		case "equipment":
			break
		default:
			return errors.New("Invalid registration status update")
	}

	user, err := h.db.GetUser(userID)
	if err != nil {
		return err
	}

	user.RegistrationStatus = status
	err = h.user.usermanager.SaveUserToDB(user)
	if err != nil {
		return err
	}

	return nil
}

// FinishRegistration function
func (h *RegistrationHandler) FinishRegistration(s *discordgo.Session, m *discordgo.MessageCreate){

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}

	err = h.perm.AddRoleToUser("Registered", user.ID, s, m, false)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}

	err = h.perm.AddRoleToUser("Crossroads", user.ID, s, m, false)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}

	err = h.user.usermanager.SaveUserToDB(user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID,"Registration complete, please enjoy your journey through *The Aether*!")
	return
}



// Attributes
func (h *RegistrationHandler) RollAttributes(s *discordgo.Session, m *discordgo.MessageCreate){

	strengthroll := RollDiceAndAdd(6, 3)
	dexterityroll := RollDiceAndAdd(6, 3)
	constituionroll := RollDiceAndAdd(6, 3)
	intelligenceroll := RollDiceAndAdd(6, 3)
	wisdomroll := RollDiceAndAdd(6, 3)
	charismaroll := RollDiceAndAdd(6, 3)


	roll := strconv.Itoa(strengthroll)
	roll = roll + " " + strconv.Itoa(dexterityroll)
	roll = roll + " " + strconv.Itoa(constituionroll)
	roll = roll + " " + strconv.Itoa(intelligenceroll)
	roll = roll + " " + strconv.Itoa(wisdomroll)
	roll = roll + " " + strconv.Itoa(charismaroll)



	attributes := "```\n"
	attributes = attributes + "Strength: " + strconv.Itoa(strengthroll) +"\n"
	attributes = attributes + "Dexterity: " + strconv.Itoa(dexterityroll) +"\n"
	attributes = attributes + "Constitution: " + strconv.Itoa(constituionroll) +"\n"
	attributes = attributes + "Intelligence: " + strconv.Itoa(intelligenceroll) +"\n"
	attributes = attributes + "Wisdom: " + strconv.Itoa(wisdomroll) +"\n"
	attributes = attributes + "Charism: " + strconv.Itoa(charismaroll) +"\n"
	attributes = attributes + "```\n"


	s.ChannelMessageSend(m.ChannelID, "Roll result: Confirm? (Yes/No):\n" + attributes)
	h.callback.Watch(h.ConfirmAttributes, GetUUIDv2(), roll, s, m)
	return
}

func (h *RegistrationHandler) ConfirmAttributes(command string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// In this handler we don't do anything with the command string, instead we grab the response from m.Content

	attributes := strings.Split(command, " ")
	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.MainConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Roll Attributes Command Cancelled")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		user, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve usermanager record: " + err.Error())
			return
		}

		user.Strength, err 			= strconv.Atoi(attributes[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}
		user.Dexterity, err			= strconv.Atoi(attributes[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}
		user.Constitution, err		= strconv.Atoi(attributes[2])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}
		user.Intelligence, err 		= strconv.Atoi(attributes[3])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}
		user.Wisdom, err			= strconv.Atoi(attributes[4])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}
		user.Charisma, err			= strconv.Atoi(attributes[5])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error with strength conversion: " + err.Error())
			return
		}

		err = h.user.usermanager.SaveUserToDB(user)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
			return
		}
	}
	if m.Content == "n" || m.Content == "no" {
		s.ChannelMessageSend(m.ChannelID, "Roll discarded, you may " +
													"re-roll with "+h.conf.MainConfig.CP+"roll-attributes.")
		return
	}

		err := h.SetRegistrationStep("attributes", m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Attributes assigned! You may now proceed with your " +
			"avatar creation by using the "+h.conf.MainConfig.CP+"pick-race command")
	return
}


// Race

func (h *RegistrationHandler) RaceInfo(s *discordgo.Session, m *discordgo.MessageCreate){

	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		racelist := GetRaceList()
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following races: \n```" + racelist +"\n```\n" )
		return
	} else {
		raceinfo := GetRaceInfo()
		raceoption := payload[0]
		if h.ValidateRaceChoice(raceoption){
			if(strings.ToLower(raceoption) == "catfolk") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[0])
				return
			}else if(strings.ToLower(raceoption) == "clockwork") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[1])
				return
			}else if(strings.ToLower(raceoption) == "dwarf") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[2])
				return
			}else if(strings.ToLower(raceoption) == "elf") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[3])
				return
			}else if(strings.ToLower(raceoption) == "halfing") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[4])
				return
			}else if(strings.ToLower(raceoption) == "half-elf") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[5])
				return
			}else if(strings.ToLower(raceoption) == "half-orc") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[6])
				return
			}else if(strings.ToLower(raceoption) == "human") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[7])
				return
			}else if(strings.ToLower(raceoption) == "kobold") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[8])
				return
			}else if(strings.ToLower(raceoption) == "gnome") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[9])
				return
			}else if(strings.ToLower(raceoption) == "orc") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[10])
				return
			}else if(strings.ToLower(raceoption) == "ratfolk") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[11])
				return
			}else if(strings.ToLower(raceoption) == "saurian") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[12])
				return
			}else if(strings.ToLower(raceoption) == "skinwalker") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+raceinfo[13])
				return
			}
		} else {
			racelist := GetRaceList()
			s.ChannelMessageSend(m.ChannelID, ":sparkles: Invalid Race Choice! You may pick from one of the following races: \n```" +
				racelist +"\n```\n" )
			return
		}
	}
}

func (h *RegistrationHandler) PickRace(s *discordgo.Session, m *discordgo.MessageCreate){

	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		racelist := "\n```Tip - Use the \"~raceinfo <race>\" command for more information about a given race \n\n" + GetRaceList()
		racelist = racelist + "\n Use \"pick-race choose <race>\" to assign an option " + "```\n"
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following races: " + racelist)
		return
	}

	for _, argument := range payload{
		argument = strings.ToLower(argument)
	}

	if len(payload) > 0 {
		raceoption := payload[0]
		if h.ValidateRaceChoice(raceoption){
			s.ChannelMessageSend(m.ChannelID, "You have chosen: " + raceoption +"\nConfirm? (Yes/No)\n")
			h.callback.Watch(h.ConfirmRace, GetUUIDv2(), raceoption, s, m)
			return
		} else {
			racelist := GetRaceList()
			s.ChannelMessageSend(m.ChannelID, ":sparkles: Invalid Race Choice! You may pick from one of the following races: \n```" +
				racelist +"\n```\n" )
			return
		}
	}
}

func (h *RegistrationHandler) ValidateRaceChoice(race string) (valid bool) {

	race = strings.ToLower(race)

	switch race {
		case "catfolk":
			return true
		case "clockwork":
			return true
		case "dwarf":
			return true
		case "elf":
			return true
		case "half-elf":
			return true
		case "half-orc":
			return true
		case "human":
			return true
		case "kobold":
			return true
		case "gnome":
			return true
		case "orc":
			return true
		case "ratfolk":
			return true
		case "saurian":
			return true
		case "skinwalker":
			return true
		default:
			return false
	}
}

func (h *RegistrationHandler) ConfirmRace(race string, s *discordgo.Session, m *discordgo.MessageCreate){

	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.MainConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Pick Race Command Cancelled")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		user, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve usermanager record: " + err.Error())
			return
		}

		user.Race = race

		err = h.user.usermanager.SaveUserToDB(user)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
			return
		}
	}
	if m.Content == "n" || m.Content == "no" {
		s.ChannelMessageSend(m.ChannelID, "Choice Cancelled.")
		return
	}

	err := h.SetRegistrationStep("race", m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Race assigned! You may now proceed with your " +
		"avatar creation by using the "+h.conf.MainConfig.CP+"pick-class command")
	return

}


// Class
func (h *RegistrationHandler) ClassInfo(s *discordgo.Session, m *discordgo.MessageCreate){

	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		classlist := GetClassList()
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following races: \n```" + classlist +"\n```\n" )
		return
	} else {
		classinfo := GetClassInfo()
		classoption := payload[0]
		if h.ValidateClassChoice(classoption){
			if(strings.ToLower(classoption) == "bard") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[0])
				return
			}else if(strings.ToLower(classoption) == "claric") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[1])
				return
			}else if(strings.ToLower(classoption) == "druid") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[2])
				return
			}else if(strings.ToLower(classoption) == "elf") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[3])
				return
			}else if(strings.ToLower(classoption) == "enchanter") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[4])
				return
			}else if(strings.ToLower(classoption) == "fighter") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[5])
				return
			}else if(strings.ToLower(classoption) == "monk") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[6])
				return
			}else if(strings.ToLower(classoption) == "necromancer") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[7])
				return
			}else if(strings.ToLower(classoption) == "ninja") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[8])
				return
			}else if(strings.ToLower(classoption) == "paladin") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[9])
				return
			}else if(strings.ToLower(classoption) == "plaguedoctor") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[10])
				return
			}else if(strings.ToLower(classoption) == "planeswalker") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[11])
				return
			}else if(strings.ToLower(classoption) == "ranger") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[12])
				return
			}else if(strings.ToLower(classoption) == "rogue") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[13])
				return
			}else if(strings.ToLower(classoption) == "shaman") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[14])
				return
			}else if(strings.ToLower(classoption) == "shaolin") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[15])
				return
			}else if(strings.ToLower(classoption) == "smuggler") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[16])
				return
			}else if(strings.ToLower(classoption) == "sorcerer") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[17])
				return
			}else if(strings.ToLower(classoption) == "wizard") {
				s.ChannelMessageSend(m.ChannelID, ":construction: "+classinfo[18])
				return
			}
		} else {
			classlist := GetClassList()
			s.ChannelMessageSend(m.ChannelID, ":sparkles: Invalid Class Choice! You may pick from one of the following classes: \n```" +
				classlist +"\n```\n" )
			return
		}
	}
}

func (h *RegistrationHandler) PickClass(s *discordgo.Session, m *discordgo.MessageCreate){

	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		classlist := "\n```Tip - Use the \"pick-class classinfo\" command for more information about a given class \n\n" + GetClassList()
		classlist = classlist + "\n Use \"pick-class choose <class>\" to assign an option " + "```\n"
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following classes: " + classlist)
		return
	}

	for _, argument := range payload{
		argument = strings.ToLower(argument)
	}

	if len(payload) > 0 {
		classoption := payload[0]
		if h.ValidateClassChoice(classoption){
			s.ChannelMessageSend(m.ChannelID, "You have chosen: " + classoption +"\nConfirm? (Yes/No)\n")
			h.callback.Watch(h.ConfirmClass, GetUUIDv2(), classoption, s, m)
			return
		} else {
			classlist := GetClassList()
			s.ChannelMessageSend(m.ChannelID, ":sparkles: Invalid Class Choice! You may pick from one of the following classes: \n```" +
				classlist +"\n```\n" )
			return
		}
	}
}

func (h *RegistrationHandler) ChooseClass(s *discordgo.Session, m *discordgo.MessageCreate){

	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		classlist := "\n```Tip - Use the \"pick-class classinfo\" command for more information about a given race \n\n" + GetClassList()
		classlist = classlist + "\n Use \"pick-race choose <race>\" to assign an option " + "```\n"
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following classes: " + classlist)
		return
	}

	for _, argument := range payload{
		argument = strings.ToLower(argument)
	}

	if len(payload) > 0 {
		classoption := payload[0]
		if h.ValidateRaceChoice(classoption){
			s.ChannelMessageSend(m.ChannelID, "You have chosen: " + classoption +"\nConfirm? (Yes/No)\n")
			h.callback.Watch(h.ConfirmClass, GetUUIDv2(), classoption, s, m)
			return
		} else {
			classlist := GetClassList()
			s.ChannelMessageSend(m.ChannelID, ":sparkles: Invalid Class Choice! You may pick from one of the following classes: \n```" +
				classlist +"\n```\n" )
			return
		}
	}

}

func (h *RegistrationHandler) ValidateClassChoice(class string) (valid bool) {

	class = strings.ToLower(class)

	switch class {

	case "barbarian":
		return true
	case "bard":
		return true
	case "cleric":
		return true
	case "druid":
		return true
	case "enchanter":
		return true
	case "fighter":
		return true
	case "monk":
		return true
	case "necromancer":
		return true
	case "ninja":
		return true
	case "paladin":
		return true
	case "plaguedoctor":
		return true

	case "planeswalker":
		return true
	case "ranger":
		return true
	case "rogue":
		return true
	case "shaman":
		return true
	case "shaolin":
		return true
	case "smuggler":
		return true
	case "sorcerer":
		return true
	case "wizard":
		return true
	default:
		return false
	}
}

func (h *RegistrationHandler) ConfirmClass(class string, s *discordgo.Session, m *discordgo.MessageCreate){

	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.MainConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Pick Class Command Cancelled")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		user, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve usermanager record: " + err.Error())
			return
		}

		user.Class = class

		err = h.user.usermanager.SaveUserToDB(user)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
			return
		}
	}
	if m.Content == "n" || m.Content == "no" {
		s.ChannelMessageSend(m.ChannelID, "Choice Cancelled.")
		return
	}

	err := h.SetRegistrationStep("class", m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Class assigned! You may now proceed with your " +
		"avatar creation by using the "+h.conf.MainConfig.CP+"pick-skills command")
	return

}


// Skills
func (h *RegistrationHandler) ChooseSkills(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmSkills(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


// Feats
func (h *RegistrationHandler) ChooseFeats(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmFeats(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


// Starter Gear
func (h *RegistrationHandler) ChooseStarterGear(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmStarterGear(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


// Misc
func (h *RegistrationHandler) ChangeMisc(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmMisc(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


// Bio
func (h *RegistrationHandler) ChangeBio(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmBio(command string, s *discordgo.Session, m *discordgo.MessageCreate){}