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
	ID 						string `storm:"id"` // primary key

	Perms 					[]uint64 // Internal Permissions - NOT Discord Roles
	RoleIDs					[]string

	// Profile stuff
	Name					string
	SkinTone				string
	Race					string
	Gender					string
	HairColor				string
	HairStyle				string
	Height					string

	Statuses 				[]string

	// Body Parts Can Have Individual States
	REye					string
	LEye 					string
	Mouth 					string
	RHand					string
	LHand					string
	RArm					string
	LArm 					string
	Head 					string
	Torso 					string
	RLeg					string
	RFoot					string
	LLeg					string
	LFoot					string

	Email 					string


	Registered				string
	RegistrationStatus		string
	RegisteredDate			time.Time


	// Related to tracking taveling
	GuildID					string `storm:"index"` // GuildID of the users current guild
	RoomID    				string `storm:"index"` // ChannelID of the users current room


	ItemsMap				[]string	// An ID pointing to the item in the database

	Strength				int
	Dexterity				int
	Constitution			int
	Intelligence			int
	Wisdom					int
	Charisma				int

	InitiativeMod			float64

	HitPoints				int64
	ExperiencePoints		int64


	Acrobatics 				int64
	Appraise				int64
	Bluff					int64
	Climb					int64
	CraftOneType			int64
	CraftOne				int64
	CraftTwoType			int64
	CraftTwo				int64
	CraftThreeType			int64
	CraftThree				int64
	Diplomacy				int64
	DisableDevice			int64
	Disguise				int64
	EscapeArtist			int64
	Fly						int64
	HandleAnimal			int64
	Heal					int64
	Intimidate				int64
	KnowledgeArcana			int64
	KnowledgeDungeoneering	int64
	KnowledgeEngineering	int64
	KnowledgeGeography		int64
	KnowledgeHistory		int64
	KnowledgeLocal			int64
	KnowledgeNature			int64
	KnowledgeNobility		int64
	KnowledgePlains			int64
	KnowledgeReligion		int64
	Linguistics				int64
	Perception				int64
	PerformOneType			string
	PerformOne				int64
	PerformTwoType			string
	PerformTwo				int64
	ProfessionOneType		int64
	ProfessionOne			int64
	ProfessionTwoType		int64
	ProfessionTwo			int64
	Ride					int64
	SenseMotive				int64
	SleightOfHand			int64
	Spellcraft				int64
	Stealth					int64
	Survival				int64
	Swim					int64
	UseMagicDevice			int64

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
		return
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
		return
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
		return u.CheckCurrentRoleList(role)
	}
}

func (u *User) CheckCurrentRoleList(roleID string) bool {
	for _, currentRole := range u.RoleIDs {
		if currentRole == roleID{
			return true
		}
	}
	return false
}

func (u *User) JoinRoleID(roleID string) {
	if u.CheckCurrentRoleList(roleID){
		return
	}

	u.RoleIDs = append(u.RoleIDs, roleID)

}


func (u *User) LeaveRoleID(roleID string) {
	if !u.CheckCurrentRoleList(roleID){
		return
	}

	u.RoleIDs = RemoveStringFromSlice(u.RoleIDs, roleID)
}