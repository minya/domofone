package lib

import (
	"github.com/minya/goutils/web"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func GetDomofoneBalance(login string, password string) (int, int, error) {
	jar := web.NewJar()
	transport := web.DefaultTransport(5000)
	client := http.Client{
		Transport: transport,
		Jar:       jar,
	}

	_, enterErr := client.Get("http://domofon-e.ru/lk/enter")
	if nil != enterErr {
		return -1, -1, enterErr
	}

	loginUrl := "http://domofon-e.ru/templates/petrunya/ajax/getData/"
	data := url.Values{}
	data.Set("get", "checkUser")
	data.Set("dg", login)
	data.Set("ps", password)
	data.Set("persData", "1")

	respGetData, errGetData := client.PostForm(loginUrl, data)
	if nil != errGetData || respGetData.StatusCode != 200 {
		return -1, -1, errGetData
	}

	petr, _ := client.Get("http://xn--e1aqefjh9f.xn--p1ai/lk/state/")
	petrBytes, _ := ioutil.ReadAll(petr.Body)
	html := string(petrBytes)

	return ParseBalance(html)
}

func ParseBalance(html string) (int, int, error) {
	reMoney, _ := regexp.Compile("<td class=\"lks03\"><b class=\"green\">(.+?),\\d\\d руб.</b>")
	match := reMoney.FindStringSubmatch(html)

	balance, errConvBal := strconv.Atoi(match[1])
	if nil != errConvBal {
		log.Printf("conv string balance to num: %v \n", errConvBal)
		return -1, -1, errConvBal
	}

	reFare, _ := regexp.Compile("<td class=\"lks03\">(.+?) руб./мес.</td>")
	matchFare := reFare.FindStringSubmatch(html)

	fare, errConvFare := strconv.Atoi(matchFare[1])
	if nil != errConvFare {
		log.Printf("conv string fare to num: %v \n", errConvFare)
		return -1, -1, errConvFare
	}
	return balance, fare, nil
}
