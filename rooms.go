package main


type Room struct {

	ID string `storm:"id"` // primary key

	ChannelID 		string 	 `storm:"index"`
	ChannelName 	string

	// Connecting Room ID's
	UpID			string
	DownID			string
	NorthID			string
	NorthEastID		string
	EastID			string
	SouthEastID		string
	SouthID			string
	SouthWestID		string
	WestID			string
	NorthWestID 	string

}


func (h *Room) AddRoomToDB(room Room, db *DBHandler) (err error) {

	return nil
}

func (h *Room) RemoveRoomFromDB(room Room, db *DBHandler) (err error) {

	return nil
}

func (h *Room) UpdateRoomFromDB(room Room, db *DBHandler) (err error) {

	return nil
}

func (h *Room) CheckIfRoomExists(room Room, db *DBHandler) (err error) {

	return nil
}