package main

import (
	"sync"
	"errors"
)

type Transfers struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type Transfer struct {

	ID              string `storm:"id"` // primary key
	TargetChannelID string `storm:"index"`
	TargetGuildID   string `storm:"index"`
	FromChannelID   string `storm:"index"`

	FromDirection	string

	UserID 			string `storm:"index"`

}


func (h *Transfers) SaveTransferToDB(transfer Transfer) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Transfers")
	err = db.Save(&transfer)
	return err
}


func (h *Transfers) RemoveTransferFromDB(transfer Transfer) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Transfers")
	err = db.DeleteStruct(&transfer)
	return err
}


func (h *Transfers) RemoveRoomByID(transferID string) (err error) {

	transfer, err := h.GetTransferByID(transferID)
	if err != nil {
		return err
	}

	err = h.RemoveTransferFromDB(transfer)
	if err != nil {
		return err
	}

	return nil
}


func (h *Transfers) GetTransferByID(roomID string) (transfer Transfer, err error) {

	transfers, err := h.GetAllTransfers()
	if err != nil{
		return transfer, err
	}

	for _, i := range transfers {
		if i.ID == roomID{
			return i, nil
		}
	}

	return transfer, errors.New("No record found")
}



// GetAllRooms function
func (h *Transfers) GetAllTransfers() (transferlist []Transfer, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Transfers")
	err = db.All(&transferlist)
	if err != nil {
		return transferlist, err
	}

	return transferlist, nil
}