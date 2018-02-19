package main

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
	// command a user should have access to.
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
		s.ChannelMessageSend(m.ChannelID, "Error finding user: " + err.Error())
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
			break;
		case "complete":
			break;
		default:
			return errors.New("Invalid registration status update")
	}

	user, err := h.db.GetUser(userID)
	if err != nil {
		return err
	}

	user.RegistrationStatus = status
	err = h.user.user.SaveUserToDB(user)
	if err != nil {
		return err
	}

	return nil
}

// FinishRegistration function
func (h *RegistrationHandler) FinishRegistration(s *discordgo.Session, m *discordgo.MessageCreate){

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	err = h.perm.AddRoleToUser("Registered", user.ID, s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}

	err = h.perm.AddRoleToUser("Crossroads", user.ID, s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not complete registration: " + err.Error())
		return
	}

	err = h.user.user.SaveUserToDB(user)
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

	// A poor way of checking the validity of the RSS url for now
	if m.Content == "" {
		s.ChannelMessageSend(m.ChannelID, "Invalid Command Received, would you like to keep this roll?")
		h.callback.Watch(h.ConfirmAttributes, GetUUIDv2(), command, s, m)
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		user, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve user record: " + err.Error())
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

		err = h.user.user.SaveUserToDB(user)
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
func (h *RegistrationHandler) PickRace(s *discordgo.Session, m *discordgo.MessageCreate){


	_, payload := SplitPayload(strings.Split(m.Content, " "))
	if len(payload) < 1{

		racelist := "\n```" +
			":bulb: Tip - Use the raceinfo command for more information about a given race \n\n" +
			"-Catfolk\n" +
			"-Clockwork\n" +
			"-Dwarf\n" +
			"-Elf\n " +
			"-Halfing\n " +
			"-Half-Elf\n " +
			"-Half-Orc\n" +
			"-Human\n" +
			"-Kobold\n" +
			"-Gnome\n" +
			"-Orc\n" +
			"-Ratfolk\n" +
			"-Saurian\n" +
			"-Skinwalker\n" +
			"```\n"
		s.ChannelMessageSend(m.ChannelID, ":sparkles: You may pick from one of the following races: " + racelist)
	}
}

func (h *RegistrationHandler) ConfirmRace(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


// Class
func (h *RegistrationHandler) ChooseClass(s *discordgo.Session, m *discordgo.MessageCreate){}

func (h *RegistrationHandler) ConfirmClass(command string, s *discordgo.Session, m *discordgo.MessageCreate){}


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