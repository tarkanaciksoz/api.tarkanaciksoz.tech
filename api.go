package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"loltracking-api/helper"
	"loltracking-api/token"

	"github.com/joho/godotenv"
)

var (
	blank       interface{}
	requestData interface{}
	tokenHash   string
)

func construct(w http.ResponseWriter, r *http.Request) interface{} {
	godotenv.Load(".env")
	helper.GlobalRequest = helper.Request{Writer: w, Request: r}
	helper.GlobalRequest.Writer.Header().Set("Content-type", "application/json")

	tokenHash = helper.GlobalRequest.Request.Header.Get("token")
	if errResp := token.CheckToken(tokenHash); errResp != blank {
		return errResp
	}

	return blank
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/getSummonerInfo", getSummonerInfo)
	http.HandleFunc("/getMatchHistoryList", getMatchHistoryList)
	http.HandleFunc("/getMatchHistory", getMatchHistory)
	http.HandleFunc("/getRankData", getRankData)
	http.ListenAndServe(":3000", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	if errResp := construct(w, r); errResp != blank {
		helper.PrintAndCleanRequest(string(errResp.([]byte)))
		return
	}
	var respon interface{}

	response := helper.SetAndGetResponse(false, "Invalid method.", respon, http.StatusBadRequest).([]byte)

	helper.PrintAndCleanRequest(string(response))
	return
}

func getRankData(w http.ResponseWriter, r *http.Request) {
	if errResp := construct(w, r); errResp != blank {
		helper.PrintAndCleanRequest(string(errResp.([]byte)))
		return
	}

	requestBody, _ := ioutil.ReadAll(helper.GlobalRequest.Request.Body)
	if len(requestBody) <= 0 {
		response := helper.SetAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["server"] == nil || data["id"] == nil {
		response := helper.SetAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}

	server := data["server"].(string)
	id := data["id"].(string)
	url := helper.GetRankDataUrl(server, id)

	cRequest, _ := http.NewRequest("GET", url, nil)
	cData, fatalResponse := helper.GetCurlData(cRequest)
	if fatalResponse != blank {
		helper.PrintAndCleanRequest(string(fatalResponse.([]byte)))
		return
	}

	response := helper.SetAndGetResponse(true, "Başarılı.", cData, 200).([]byte)

	helper.PrintAndCleanRequest(string(response))
	return

}

func getSummonerInfo(w http.ResponseWriter, r *http.Request) {
	if errResp := construct(w, r); errResp != blank {
		helper.PrintAndCleanRequest(string(errResp.([]byte)))
		return
	}

	requestBody, _ := ioutil.ReadAll(helper.GlobalRequest.Request.Body)
	if len(requestBody) <= 0 {
		response := helper.SetAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["server"] == nil || data["userName"] == nil {
		response := helper.SetAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}

	server := data["server"].(string)
	userName := data["userName"].(string)
	url := helper.GetSummonerProfileUrl(server, userName)

	cRequest, _ := http.NewRequest("GET", url, nil)
	cData, fatalResponse := helper.GetCurlData(cRequest)
	if fatalResponse != blank {
		helper.PrintAndCleanRequest(string(fatalResponse.([]byte)))
		return
	}

	response := helper.SetAndGetResponse(true, "Başarılı.", cData, 200).([]byte)

	helper.PrintAndCleanRequest(string(response))
	return
}

func getMatchHistoryList(w http.ResponseWriter, r *http.Request) {
	if errResp := construct(w, r); errResp != blank {
		helper.PrintAndCleanRequest(string(errResp.([]byte)))
		return
	}

	var (
		url       string
		puuId     string
		queue     string = ""
		queueType string = ""
		offset    string = "0"
		limit     string = "10"
		cData     interface{}
	)

	requestBody, _ := ioutil.ReadAll(helper.GlobalRequest.Request.Body)
	if len(requestBody) <= 0 {
		response := helper.SetAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["puuId"] == nil {
		response := helper.SetAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}

	puuId = data["puuId"].(string)

	if data["queue"] != nil && len(data["queue"].(string)) > 0 {
		queue = data["queue"].(string)
	}
	if data["queueType"] != nil && len(data["queueType"].(string)) > 0 {
		queueType = data["queueType"].(string)
	}
	if data["offset"] != nil && len(data["offset"].(string)) > 0 {
		offset = data["offset"].(string)
	}
	if data["limit"] != nil && len(data["limit"].(string)) > 0 {
		limit = data["limit"].(string)
	}

	url = helper.GetMatchHistorListUrl(puuId, queue, queueType, offset, limit)

	cRequest, err := http.NewRequest("GET", url, nil)
	if errResponse := helper.ErrorResponse(err); errResponse != blank {
		helper.PrintAndCleanRequest(string(errResponse.([]byte)))
		return
	}
	cData, fatalResponse := helper.GetCurlData(cRequest)
	if fatalResponse != blank {
		helper.PrintAndCleanRequest(string(fatalResponse.([]byte)))
		return
	}

	response := helper.SetAndGetResponse(true, "Başarılı.", cData, 200).([]byte)

	helper.PrintAndCleanRequest(string(response))
	return
}

func getMatchHistory(w http.ResponseWriter, r *http.Request) {
	if errResp := construct(w, r); errResp != blank {
		helper.PrintAndCleanRequest(string(errResp.([]byte)))
		return
	}

	requestBody, _ := ioutil.ReadAll(helper.GlobalRequest.Request.Body)
	if len(requestBody) <= 0 {
		response := helper.SetAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}

	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["idList"] == nil {
		response := helper.SetAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		helper.PrintAndCleanRequest(string(response))
		return
	}

	var url string
	var matchHistoryArr []interface{}
	list := data["idList"].([]interface{})
	for _, matchId := range list {
		url = helper.GetMatchHistoryUrl(matchId.(string))
		cRequest, err := http.NewRequest("GET", url, nil)
		if errResponse := helper.ErrorResponse(err); errResponse != blank {
			helper.PrintAndCleanRequest(string(errResponse.([]byte)))
			return
		}
		cData, fatalResponse := helper.GetCurlData(cRequest)
		if fatalResponse != blank {
			helper.PrintAndCleanRequest(string(fatalResponse.([]byte)))
			return
		}

		matchHistoryArr = append(matchHistoryArr, cData)
	}

	response := helper.SetAndGetResponse(true, "Başarılı.", matchHistoryArr, 200).([]byte)

	helper.PrintAndCleanRequest(string(response))
	return
}
