package main

import (
	"flag"
	"log"
	"os"

	"github.com/minya/domofone/lib"
	"github.com/minya/gopushover"
	"github.com/minya/goutils/config"
)

var logPath string

func init() {
	const (
		defaultLogPath = "domofone.log"
	)
	flag.StringVar(&logPath, "logpath", defaultLogPath, "Path to write logs")
}

func main() {
	flag.Parse()

	setUpLogger()
	log.Printf("Start\n")

	var settings settings
	settingsErr := config.UnmarshalJson(&settings, "~/.domofone/settings.json")
	if nil != settingsErr {
		log.Fatalf("read settings: %v \n", settingsErr)
	}

	balance, fare, errParse := lib.GetDomofoneBalance(settings.DomofonELogin, settings.DomofonEPassword)
	if nil != errParse {
		log.Fatalf("Unable to parse html: %v \n", errParse)
	}

	log.Printf("Values obtained. Balance: %v, fare: %v\n", balance, fare)

	lastAction := getLastAction()
	log.Printf("Last action was %v\n", lastAction)
	if balance < 2*fare {
		log.Printf("Balance is low\n")
		if lastAction != "notify" {
			poSettings, poErr := gopushover.ReadSettings("~/.domofone/pushover.json")
			if nil != poErr {
				log.Fatalf("pushover settings: %v \n", poErr)
			}

			pushRes, pushErr := gopushover.SendMessage(
				poSettings.Token, poSettings.User, "Domofon-e low balance", "Your balance is low")
			if nil != pushErr {
				log.Fatalf("Can't push: %v\n")
			}

			if nil != pushRes.Errors && len(pushRes.Errors) > 0 {
				log.Fatalf("%v from pushover\n", pushRes.Errors)
			}

			lastAction = "notify"
		}
	} else {
		log.Printf("Balance is OK\n")
		if lastAction == "notify" {
			lastAction = "pass"
		}
	}
	log.Printf("Last action is being set to %v\n", lastAction)
	setLastAction(lastAction)
}

func setUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

func getLastAction() string {
	var state state
	errDeser := config.UnmarshalJson(&state, "~/.domofone/state.json")
	if nil != errDeser {
		return "pass"
	}
	return state.LastAction
	return "pass"
}

func setLastAction(value string) error {
	state := state{}
	state.LastAction = value
	errWrite := config.MarshalJson(state, "~/.domofone/state.json")
	if nil != errWrite {
		return errWrite
	}
	return nil
}

type state struct {
	LastAction string
}

type settings struct {
	DomofonELogin    string
	DomofonEPassword string
}
