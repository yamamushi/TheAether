package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"time"
)

// Config struct
type Config struct {
	MainConfig mainConfig `toml:"main"`
	BankConfig bankConfig `toml:"bank"`
}

// discordConfig struct
type mainConfig struct {

	// Command Prefix
	Token          string        `toml:"bot_token"`
	ClusterOwnerID string        `toml:"cluster_owner_id"`
	CentralGuildID string        `toml:"central_Server_id"`
	LobbyChannelID string        `toml:"lobby_channel_id"`
	CP             string        `toml:"default_command_prefix"`
	Playing        string        `toml:"default_now_playing"`
	Notifications  time.Duration `toml:"notifications_update_timeout"`
	PerPageCount   int           `toml:"per_page_count"`
	LuaTimeout     int           `toml:"lua_timeout"`
	Profiler       bool          `toml:"enable_profiler"`
	DBFile         string        `toml:"dbfilename"`
}

// bankConfig struct
type bankConfig struct {
	BankName               string `toml:"bank_name"`
	BankURL                string `toml:"bank_url"`
	BankIconURL            string `toml:"bank_icon_url"`
	Pin                    string `toml:"bank_pin"`
	Reset                  bool   `toml:"reset_bank"`
	SeedWallet             int    `toml:"starting_bank_wallet_value"`
	SeedUserAccountBalance int    `toml:"starting_user_account_value"`
	SeedUserWalletBalance  int    `toml:"starting_user_wallet_value"`
	BankMenuSlogan         string `toml:"bank_menu_slogan"`
}

// ReadConfig function
func ReadConfig(path string) (config Config, err error) {

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}
