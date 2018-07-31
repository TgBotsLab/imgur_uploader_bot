package main

import (
	"github.com/integrii/flaggy"
	"github.com/tgbotslab/imgur-uploader-bot"
)

const version = "0.1"

var token, clientID, tmpDir, descPhoto string

func init() {
	flaggy.SetName("Imgur Uploader Bot")
	flaggy.SetDescription("The bot uploads photos to Imgur")

	flaggy.String(&token, "t", "token", "telegram bot token")
	flaggy.String(&clientID, "c", "client-id", "imgur client id")
	flaggy.String(&tmpDir, "d", "tmp-dir", "temporary folder to store photos")
	flaggy.String(&descPhoto, "dp", "desc-photo", "description to uploaded photos")

	flaggy.SetVersion(version)
	flaggy.Parse()
}

func main() {
	bot.Init(token, clientID, tmpDir, descPhoto)
}
