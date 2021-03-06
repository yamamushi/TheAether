package main

import (
	"fmt"
	"github.com/asdine/storm"
)

// DBHandler struct
type DBHandler struct {
	rawdb *storm.DB
	conf  *Config
}

// FirstTimeSetup function
func (h *DBHandler) FirstTimeSetup() error {

	//	usermanager.ID = h.conf.DiscordConfig.AdminID

	//_ := h.rawdb.From("Users")

	/*	err := db.One("ID", h.conf.DiscordConfig.AdminID, &usermanager)
		if err != nil {
			fmt.Println("Running first time db config")
			walletdb := db.From("Wallets")
			usermanager.SetRole("owner")
			err := db.Save(&usermanager)
			if err != nil {
				fmt.Println("error saving owner")
				return err
			}

			wallet := Wallet{Account: h.conf.DiscordConfig.AdminID, Balance: 10000}
			err = walletdb.Save(&wallet)
			if err != nil {
				fmt.Println("error saving wallet")
				return err
			}

			if usermanager.Owner {
				err = db.One("ID", h.conf.DiscordConfig.AdminID, &usermanager)
				if err != nil {
					fmt.Println("Could not retrieve data from the database, something went wrong!")
					return err
				}
				fmt.Println("Owner ID: " + usermanager.ID)
				fmt.Println("Database has been configured")
				return nil
			}
		}
	*/
	return nil
}

// Insert function
func (h *DBHandler) Insert(object interface{}) error {

	err := h.rawdb.Save(object)
	if err != nil {
		fmt.Println("Could not insert object: ", err.Error())
		return err
	}

	return nil
}

// Find function
func (h *DBHandler) Find(first string, second string, object interface{}) error {

	err := h.rawdb.One(first, second, object)
	if err != nil {
		return err
	}
	return nil
}

// Update function
func (h *DBHandler) Update(object interface{}) error {
	err := h.rawdb.Update(object)
	if err != nil {
		return err
	}
	return nil
}

// GetUser function
func (h *DBHandler) GetUser(uid string) (user User, err error) {

	userdb := h.rawdb.From("Users")
	err = userdb.One("ID", uid, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
