package bot

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-scan"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	endpoint = "https://api.imgur.com/3/image"

	howToUse           = "Just send a photo to me."
	somethingWentWrong = "Sorry, but something went wrong. Try again later."
	gotPhoto           = "Okay. We have got your photo. Wait, please."
	photoUploaded      = "Your photo was successfully uploaded to Imgur."
)

func Init(token, clientID, tmpDir, descPhoto string) {
	log.Println("Imgur Uploader Bot is started")

	// Connect to Telegram bot API
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	// It is a shortcut to send a reply
	reply := func(b *tb.Bot, m *tb.Message, msg string) {
		b.Send(m.Sender, msg)
	}

	// Here we handle uploaded photos and process them
	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		// Save a uploaded photo to our temporary folder
		f, err := b.FileByID(m.Photo.File.FileID)
		if err != nil {
			log.Println(errors.Wrap(err, "cannot get a file by its ID"))
			reply(b, m, somethingWentWrong)
			return
		}

		// Download the photo
		filename := tmpDir + randName(f.FileID) + filepath.Ext(f.FilePath)
		b.Download(&f, filename)
		reply(b, m, gotPhoto)

		// Upload a photo to Imgur
		bFile, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(errors.Wrap(err, "cannot read a file"))
			reply(b, m, somethingWentWrong)
			return
		}

		// Uploading to Imgur was taken from
		// https://github.com/mattn/imgur/blob/master/imgur.go
		params := url.Values{
			"image":       {base64.StdEncoding.EncodeToString(bFile)},
			"description": {descPhoto},
		}

		req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
		if err != nil {
			log.Println(errors.Wrap(err, "cannot make a request"))
			reply(b, m, somethingWentWrong)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Client-ID "+clientID)

		var res *http.Response
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println(errors.Wrap(err, "cannot make a request"))
			reply(b, m, somethingWentWrong)
			return
		}

		if res.StatusCode != 200 {
			var message string
			err = scan.ScanJSON(res.Body, "data/error", &message)
			if err != nil {
				message = res.Status
				log.Println(errors.Wrap(err, "cannot get an error"), message)
			}
			reply(b, m, somethingWentWrong)
			return
		}
		defer res.Body.Close()

		var link string
		err = scan.ScanJSON(res.Body, "data/link", &link)
		if err != nil {
			log.Println(errors.Wrap(err, "cannot get a link"))
			reply(b, m, somethingWentWrong)
			return
		}

		reply(b, m, photoUploaded)
		reply(b, m, link)

		log.Println("A photo was uploaded: ", link)

		// Remove a temp file
		err = os.Remove(filename)
		if err != nil {
			log.Println(errors.Wrap(err, "cannot delete a temp file"))
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		reply(b, m, howToUse)
	})

	b.Handle("/start", func(m *tb.Message) {
		reply(b, m, howToUse)
	})

	b.Start()
}

// randName generates a random string from s
func randName(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
