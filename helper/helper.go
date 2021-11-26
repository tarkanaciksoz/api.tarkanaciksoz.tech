package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Request struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

type Response struct {
	Success bool        `"json:success"`
	Message string      `"json:message"`
	Data    interface{} `"json:data"`
	Code    int         `"json:code"`
}

var HttpStatusForbidden = http.StatusForbidden

var (
	blank          interface{}
	GlobalResponse Response
	GlobalRequest  Request
)

func GetCurlData(request *http.Request) interface{} {
	var data interface{}
	response, _ := http.DefaultClient.Do(request)

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	json.Unmarshal(body, &data)

	return data
}

func GetUrlWithApiKey(url string) string {
	return url + os.Getenv("API_KEY")
}

func GetSummonerProfileUrl(server string, userName string) string {
	summonerUrl := GetUrlWithApiKey("https://" + server + ".api.riotgames.com/lol/summoner/v4/summoners/by-name/" + userName + "?api_key=")

	return summonerUrl
}

func GetMatchHistoryUrl(matchId string) string {
	return GetUrlWithApiKey("https://europe.api.riotgames.com/lol/match/v5/matches/" + matchId + "?api_key=")
}

func GetMatchHistorListUrl(puuId string, queue string, queueType string, offset string, limit string) string {
	var url string = "https://europe.api.riotgames.com/lol/match/v5/matches/by-puuid/" + puuId + "/ids?"

	if q := queue; q != "" {
		url += "queue=" + queue + "&"
	}
	if qT := queueType; qT != "" {
		url += "type=" + queueType + "&"
	}

	url += "start=" + offset + "&count=" + limit + "&api_key="
	return GetUrlWithApiKey(url)
}

func SetAndGetResponse(success bool, message string, data interface{}, code int) interface{} {
	var response interface{}
	GlobalResponse = Response{Success: success, Message: message, Data: data, Code: code}

	successResponse, err := json.Marshal(GlobalResponse)
	fatalResponse := ErrorResponse(err)

	if response = successResponse; fatalResponse != blank {
		response, _ = json.Marshal(fatalResponse)
	}

	return response
}

func ErrorResponse(err error) interface{} {
	if err != nil {
		fatalResponse := SetAndGetResponse(false, err.Error(), nil, http.StatusBadRequest)
		return fatalResponse
	}
	return blank
}

func PrintAndCleanRequest(printResponse string) {
	fmt.Fprint(GlobalRequest.Writer, printResponse)
	GlobalRequest = Request{}

	return
}
