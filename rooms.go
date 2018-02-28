package main

import (
	"errors"
	"sync"
)

// Rooms struct
type Rooms struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// Room struct
type Room struct {
	ID string `storm:"id"` // primary key

	Name    string
	GuildID string
	Type    string

	Topic string // For recording the discord room topic

	ParentID   string
	ParentName string

	TravelRoleID      string
	AdditionalRoleIDs []string
	UserIDs           []string
	/*
		These are likely to change, but here's generally what is needed in the roleID slice
		0 - Default Travel ID
		1 - Region Role ID
		2 - Quest Role ID
		3 - Spell Role ID
		4 - Blessing Role ID
		5 - Override Role ID
	*/

	GuildTransferInvite string
	TransferRoomID      string

	// Connecting Room ID's
	UpID       string
	UpItemID   []string
	UpScriptID string

	DownID       string
	DownItemID   []string
	DownScriptID string

	NorthID       string
	NorthItemID   []string
	NorthScriptID string

	NorthEastID       string
	NorthEastItemID   []string
	NorthEastScriptID string

	EastID       string
	EastItemID   []string
	EastScriptID string

	SouthEastID       string
	SouthEastItemID   []string
	SouthEastScriptID string

	SouthID       string
	SouthItemID   []string
	SouthScriptID string

	SouthWestID       string
	SouthWestItemID   []string
	SouthWestScriptID string

	WestID       string
	WestItemID   []string
	WestScriptID string

	NorthWestID       string
	NorthWestItemID   []string
	NorthWestScriptID string

	Items []string
	NPC   []string

	Description string
}

// SaveRoomToDB function
func (h *Rooms) SaveRoomToDB(room Room) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Rooms")
	err = db.Save(&room)
	return err
}

// RemoveRoomFromDB function
func (h *Rooms) RemoveRoomFromDB(room Room) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Rooms")
	err = db.DeleteStruct(&room)
	return err
}

// RemoveRoomByID function
func (h *Rooms) RemoveRoomByID(roomID string) (err error) {

	room, err := h.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	err = h.RemoveRoomFromDB(room)
	if err != nil {
		return err
	}

	return nil
}

// GetRoomByID function
func (h *Rooms) GetRoomByID(roomID string) (room Room, err error) {

	rooms, err := h.GetAllRooms()
	if err != nil {
		return room, err
	}

	for _, record := range rooms {

		if roomID == record.ID {
			return record, nil
		}
	}
	return room, errors.New("No record found")
}

// GetRoomByName function
func (h *Rooms) GetRoomByName(roomname string, guildID string) (room Room, err error) {

	rooms, err := h.GetAllRooms()
	if err != nil {
		return room, err
	}

	for _, i := range rooms {
		if i.Name == roomname && i.GuildID == guildID {
			return i, nil
		}
	}

	return room, errors.New("No record found")
}

// GetAllRooms function
func (h *Rooms) GetAllRooms() (roomlist []Room, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Rooms")
	err = db.All(&roomlist)
	if err != nil {
		return roomlist, err
	}

	return roomlist, nil
}

// IsRoomLinkedTo function
func (h *Rooms) IsRoomLinkedTo(roomID string, checklink string) (linked bool, err error) {

	room, err := h.GetRoomByID(roomID)
	if err != nil {
		return false, err
	}
	if room.NorthID == checklink {
		return true, nil
	}
	if room.NorthEastID == checklink {
		return true, nil
	}
	if room.EastID == checklink {
		return true, nil
	}
	if room.SouthEastID == checklink {
		return true, nil
	}
	if room.SouthID == checklink {
		return true, nil
	}
	if room.SouthWestID == checklink {
		return true, nil
	}
	if room.WestID == checklink {
		return true, nil
	}
	if room.NorthWestID == checklink {
		return true, nil
	}
	if room.UpID == checklink {
		return true, nil
	}
	if room.DownID == checklink {
		return true, nil
	}
	return false, nil
}
