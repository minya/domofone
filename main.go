package main

import (
	"encoding/json"
	"flag"
	"github.com/minya/gopushover"
	"github.com/minya/goutils/web"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
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

	SetUpLogger()
	log.Printf("Start\n")
	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport: transport,
		Jar:       jar,
	}

	_, enterErr := client.Get("http://domofon-e.ru/lk/enter")
	if nil != enterErr {
		log.Fatalf("site is down: %v\n", enterErr)
	}

	user, _ := user.Current()
	settingsPath := path.Join(user.HomeDir, ".domofone/settings.json")

	var settings Settings
	settingsBin, settingsErr := ioutil.ReadFile(settingsPath)
	if nil != settingsErr {
		log.Fatalf("read settings: %v \n", settingsErr)
	}

	json.Unmarshal(settingsBin, &settings)
	loginUrl := "http://domofon-e.ru/templates/petrunya/ajax/getData/"
	data := url.Values{}
	data.Set("get", "checkUser")
	data.Set("dg", settings.DomofonELogin)
	data.Set("ps", settings.DomofonEPassword)

	respGetData, errGetData := client.PostForm(loginUrl, data)
	if nil != errGetData || respGetData.StatusCode != 200 {
		log.Fatalf("getData: %v \n", errGetData)
	}

	petr, _ := client.Get("http://xn--e1aqefjh9f.xn--p1ai/lk/state/")
	petrBytes, _ := ioutil.ReadAll(petr.Body)
	html := string(petrBytes)

	reMoney, _ := regexp.Compile("<td class=\"lks03\"><b class=\"green\">(.+?) руб.</b>")
	match := reMoney.FindStringSubmatch(html)

	balance, errConvBal := strconv.Atoi(match[1])
	if nil != errConvBal {
		log.Fatalf("conv string balance to num: %v \n", errConvBal)
	}

	reFare, _ := regexp.Compile("<td class=\"lks03\">(.+?) руб./мес.</td>")
	matchFare := reFare.FindStringSubmatch(html)

	fare, errConvFare := strconv.Atoi(matchFare[1])
	if nil != errConvFare {
		log.Fatalf("conv string fare to num: %v \n", errConvFare)
	}

	log.Printf("Values obtained. Balance: %v, fare: %v\n", balance, fare)

	lastAction := GetLastAction()
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
	SetLastAction(lastAction)
}

func SetUpLogger() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}

type Settings struct {
	DomofonELogin    string
	DomofonEPassword string
}

func GetLastAction() string {
	content, errRead := ioutil.ReadFile(ExpandUserHome("~/.domofone/state.json"))
	if nil != errRead {
		return "pass"
	}
	var state State
	errDeser := json.Unmarshal(content, &state)
	if nil != errDeser {
		return "pass"
	}
	return state.LastAction
	return "pass"
}

func SetLastAction(value string) error {
	state := State{}
	state.LastAction = value
	content, errSer := json.Marshal(state)
	if nil != errSer {
		return errSer
	}
	errWrite := ioutil.WriteFile(ExpandUserHome("~/.domofone/state.json"), content, 0660)
	if nil != errWrite {
		return errWrite
	}
	return nil
}

func ExpandUserHome(spath string) string {
	if strings.Index(spath, "~/") != 0 {
		return spath
	}
	user, _ := user.Current()
	return path.Join(user.HomeDir, strings.TrimLeft(spath, "~/"))
}

type State struct {
	LastAction string
}
