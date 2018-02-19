package main

import (
	"time"
	"sync"
	"errors"
)


type UserManager struct {
	db          *DBHandler
	querylocker sync.RWMutex
}
// User struct
type User struct {
	ID string `storm:"id"` // primary key

	Perms 		[]uint64 // Internal Permissions - NOT Discord Roles
	Roles		[]string

	// Profile stuff
	Name		string
	SkinTone	string
	Gender		string
	HairColor	string
	HairStyle	string
	Height		string

	Statuses 	[]string

	// Body Parts Can Have Individual States
	REye		string
	LEye 		string
	Mouth 		string
	RHand		string
	LHand		string
	RArm		string
	LArm 		string
	Head 		string
	Torso 		string
	RLeg		string
	RFoot		string
	LLeg		string
	LFoot		string

	Email 		string


	Registered				string
	RegistrationStatus		string
	RegisteredDate			time.Time


	// Related to tracking taveling
	GuildID		string `storm:"index"` // GuildID of the users current guild
	RoomID    	string `storm:"index"` // ChannelID of the users current room


	ItemsMap	[]string	// An ID pointing to the item in the database

	Strength	float64
	Dexterity	float64
	Constitution	float64
	Intelligence	float64
	Wisdom		float64
	Charisma	float64

	InitiativeMod	float64

	HitPoints	float64
	ExperiencePoints		int64


	Acrobatics float64
	Appraise	float64
	Bluff	float64
	Climb	float64
	CraftOneType	float64
	CraftOne	float64
	CraftTwoType	float64
	CraftTwo	float64
	CraftThreeType	float64
	CraftThree	float64
	Diplomacy	float64
	DisableDevice	float64
	Disguise	float64
	EscapeArtist	float64
	Fly	float64
	HandleAnimal	float64
	Heal	float64
	Intimidate	float64
	KnowledgeArcana	float64
	KnowledgeDungeoneering	float64
	KnowledgeEngineering	float64
	KnowledgeGeography	float64
	KnowledgeHistory	float64
	KnowledgeLocal	float64
	KnowledgeNature	float64
	KnowledgeNobility	float64
	KnowledgePlains	float64
	KnowledgeReligion	float64
	Linguistics	float64
	Perception	float64
	PerformOneType			string
	PerformOne	float64
	PerformTwoType			string
	PerformTwo	float64
	ProfessionOneType	float64
	ProfessionOne	float64
	ProfessionTwoType	float64
	ProfessionTwo	float64
	Ride	float64
	SenseMotive	float64
	SleightOfHand	float64
	Spellcraft	float64
	Stealth	float64
	Survival	float64
	Swim	float64
	UseMagicDevice	float64

	Spellbook				string // An ID of the spellbook in the database

	// Money
	CopperPieces			int64
	SilverPieces			int64
	GoldPieces				int64
	PlatinumPieces			int64

}




func (h *UserManager) SaveUserToDB(user User) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Users")
	err = db.Save(&user)
	return err
}

func (h *UserManager) RemoveUserFromDB(user User) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Users")
	err = db.DeleteStruct(&user)
	return err
}

func (h *UserManager) RemoveUserByID(userID string) (err error) {

	room, err := h.GetUserByID(userID)
	if err != nil {
		return err
	}

	err = h.RemoveUserFromDB(room)
	if err != nil {
		return err
	}

	return nil
}

func (h *UserManager) GetUserByID(userID string) (user User, err error) {

	users, err := h.GetAllUsers()
	if err != nil{
		return user, err
	}

	for _, i := range users {
		if i.ID == userID{
			return i, nil
		}
	}

	return user, errors.New("No record found")
}

func (h *UserManager) GetUserByName(username string, guildID string) (user User, err error) {

	users, err := h.GetAllUsers()
	if err != nil{
		return user, err
	}

	for _, i := range users {
		if i.Name == username && i.GuildID == guildID{
			return i, nil
		}
	}

	return user, errors.New("No record found")
}

func (h *UserManager) GetAllUsers() (userlist []User, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Users")
	err = db.All(&userlist)
	if err != nil {
		return userlist, err
	}

	return userlist, nil
}



// Init function
func (u *User) Init() {
	ClearRoles(u)
	PlayerRole(u)
}

// SetRole function
func (u *User) SetRole(role string) {

	switch role {

	case "owner":
		OwnerRole(u)

	case "admin":
		AdminRole(u)

	case "smoderator":
		SModeratorRole(u)

	case "moderator":
		ModeratorRole(u)

	case "builder":
		BuilderRole(u)

	case "writer":
		WriterRole(u)

	case "scripter":
		ScripterRole(u)

	case "architect":
		ArchitectRole(u)

	case "player":
		PlayerRole(u)

	case "clear":
		ClearRoles(u)

	default:
		u.JoinRole(role)
	}
}

// RemoveRole function
func (u *User) RemoveRole(role string) {

	switch role {

	case "owner":
		SetBit(&u.Perms, 60)

	case "admin":
		SetBit(&u.Perms, 59)

	case "smoderator":
		SetBit(&u.Perms, 58)

	case "moderator":
		SetBit(&u.Perms, 57)

	case "builder":
		SetBit(&u.Perms, 56)

	case "writer":
		SetBit(&u.Perms, 55)

	case "scripter":
		SetBit(&u.Perms, 54)

	case "architect":
		SetBit(&u.Perms, 53)

	case "player":
		SetBit(&u.Perms, 10)

	default:
		u.LeaveRole(role)
	}
}

// CheckRole function
func (u *User) CheckRole(role string) bool {

	switch role {

	case "owner":
		return IsBitSet(&u.Perms, 60)

	case "admin":
		return IsBitSet(&u.Perms, 59)

	case "smoderator":
		return IsBitSet(&u.Perms, 58)

	case "moderator":
		return IsBitSet(&u.Perms, 57)

	case "builder":
		return IsBitSet(&u.Perms, 56)

	case "writer":
		return IsBitSet(&u.Perms, 55)

	case "scripter":
		return IsBitSet(&u.Perms, 54)

	case "architect":
		return IsBitSet(&u.Perms, 53)

	case "player":
		return IsBitSet(&u.Perms, 10)

	default:
		return u.CheckDiscordRole(role)
	}
}

func (u *User) CheckDiscordRole(rolename string) bool {

	for _, role := range u.Roles {

		if role == rolename{
			return true
		}
	}
	return false
}

func (u *User) JoinRole(rolename string) {
	if u.CheckDiscordRole(rolename){
		return
	}

	u.Roles = append(u.Roles, rolename)

}


func (u *User) LeaveRole(rolename string) {
	if !u.CheckDiscordRole(rolename){
		return
	}

	u.Roles = RemoveStringFromSlice(u.Roles, rolename)
}