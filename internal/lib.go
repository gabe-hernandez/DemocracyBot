package internal

import (
	"strings"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"
	"github.com/bwmarrin/discordgo"
)

var usernameToId map[string]string
const Threshold int = 1
const invalidCommand = "I'm afraid I can't do that...(yet)"

var tokenFile = "bot.key"
var DefaultVoteTime, _ = time.ParseDuration("10s")

func StrInSlice(str string, slice []string) bool {
	for _, v := range slice {
		if strings.Compare(v, str) == 0 {
			return true
		}
	}
	return false
}


func IsAdmin(user string) bool {
	switch user {
	case
		"QuantumQuip",
		"picthebear",
		"MetalKnuckles":
		return true
	}
	return false
}

func GetUserIDfromString(s *discordgo.Session, guildID string, user string) (string, error) {
	if usernameToId == nil {
		usernameToId = make(map[string]string)
		members, err := s.GuildMembers(guildID, "", 100)

		if err != nil {
			return "", err
		}

		for _, mem := range members {
			usernameToId[mem.User.Username] = mem.User.ID

			if len(mem.Nick) > 0 {
				usernameToId[mem.Nick] = mem.User.ID
			}
		}
	}

	ID, ok := usernameToId[user]

	if !ok {
		return "", errors.New("Username not found")
	}

	return ID, nil
}

func ReadKeyFile() string {
	file, err := os.Open(tokenFile)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	tokenSlice, err := ioutil.ReadAll(file)
	return string(tokenSlice)
}