package main

import (
	"encoding/json"
	"fmt"
	"github.com/minya/goutils/web"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"path"
	"regexp"
)

func main() {
	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport: transport,
		Jar:       jar,
	}

	_, enterErr := client.Get("http://domofon-e.ru/lk/enter")
	if nil != enterErr {
		fmt.Printf("Zalupa: %v", enterErr)
		return
	}

	user, _ := user.Current()
	settingsPath := path.Join(user.HomeDir, ".domofone/settings.json")

	var settings Settings
	settingsBin, settingsErr := ioutil.ReadFile(settingsPath)
	if nil != settingsErr {
		fmt.Printf("Zalupa (read settings): %v \n", settingsErr)
		return
	}

	json.Unmarshal(settingsBin, &settings)
	loginUrl := "http://domofon-e.ru/templates/petrunya/ajax/getData/"
	data := url.Values{}
	data.Set("get", "checkUser")
	data.Set("dg", settings.DomofonELogin)
	data.Set("ps", settings.DomofonEPassword)

	respGetData, errGetData := client.PostForm(loginUrl, data)
	if nil != errGetData || respGetData.StatusCode != 200 {
		fmt.Printf("Zalupa (getData): %v \n", enterErr)
		return
	}

	petr, _ := client.Get("http://xn--e1aqefjh9f.xn--p1ai/lk/state/")
	petrBytes, _ := ioutil.ReadAll(petr.Body)
	html := string(petrBytes)

	reMoney, _ := regexp.Compile("<td class=\"lks03\"><b class=\"green\">(.+?) руб.</b>")
	match := reMoney.FindStringSubmatch(html)

	fmt.Printf(match[1] + "\n")

	reFare, _ := regexp.Compile("<td class=\"lks03\">(.+?) руб./мес.</td>")
	matchFare := reFare.FindStringSubmatch(html)

	fmt.Printf(matchFare[1] + "\n")
}

type Settings struct {
	DomofonELogin    string
	DomofonEPassword string
}
