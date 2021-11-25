package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Globals interface {
	getWriter() http.ResponseWriter
	getRequest() *http.Request
	getResponse() Response
}

type TokenRow struct {
	id       int
	token    string
	isActive int
}

type Request struct {
	writer  http.ResponseWriter
	request *http.Request
}

type Response struct {
	Success bool        `"json:success"`
	Message string      `"json:message"`
	Data    interface{} `"json:data"`
	Code    int         `"json:code"`
}

var (
	db             *sqlx.DB
	token          string
	globalRequest  Request
	globalResponse Response
	blank          interface{}
	summonerUrl    string
	requestData    interface{}
	apiKey         string
)

func construct(w http.ResponseWriter, r *http.Request) bool {
	godotenv.Load(".env")
	apiKey = os.Getenv("API_KEY")
	globalRequest = Request{writer: w, request: r}
	globalRequest.getWriter().Header().Set("Content-type", "application/json")

	token = globalRequest.getRequest().Header.Get("token")
	if errResp := checkToken(token); errResp != blank {
		fmt.Fprint(globalRequest.getWriter(), string(errResp.([]byte)))
		return false
	}

	return true
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/getSummonerInfo", getSummonerInfo)
	http.HandleFunc("/getMatchHistoryList", getMatchHistoryList)
	http.HandleFunc("/getMatchHistory", getMatchHistory)
	http.HandleFunc("/riot.txt", riotTxt)
	http.ListenAndServe(":3000", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}
	var respon interface{}

	response := setAndGetResponse(false, "Invalid method.", respon, http.StatusBadRequest).([]byte)

	fmt.Fprint(globalRequest.getWriter(), string(response))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func getSummonerInfo(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}

	requestBody, _ := ioutil.ReadAll(globalRequest.getRequest().Body)
	if len(requestBody) <= 0 {
		response := setAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["server"] == nil || data["userName"] == nil {
		response := setAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}

	server := data["server"].(string)
	userName := data["userName"].(string)
	url := getSummonerProfileUrl(server, userName)

	cRequest, _ := http.NewRequest("GET", url, nil)
	cData := getCurlData(cRequest)

	response := setAndGetResponse(true, "Başarılı.", cData, 200).([]byte)

	fmt.Fprint(globalRequest.getWriter(), string(response))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func getMatchHistoryList(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		w = nil
		r = nil
		globalRequest = Request{}
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

	requestBody, _ := ioutil.ReadAll(globalRequest.getRequest().Body)
	if len(requestBody) <= 0 {
		response := setAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["puuId"] == nil {
		response := setAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
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

	url = getMatchHistorListUrl(puuId, queue, queueType, offset, limit)

	cRequest, err := http.NewRequest("GET", url, nil)
	if errResponse := errorResponse(err); errResponse != blank {
		fmt.Fprint(globalRequest.getWriter(), string(errResponse.([]byte)))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}
	cData = getCurlData(cRequest)

	response := setAndGetResponse(true, "Başarılı.", cData, 200).([]byte)

	fmt.Fprint(globalRequest.getWriter(), string(response))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func getMatchHistory(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}

	requestBody, _ := ioutil.ReadAll(globalRequest.getRequest().Body)
	if len(requestBody) <= 0 {
		response := setAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}

	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["idList"] == nil {
		response := setAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		w = nil
		r = nil
		globalRequest = Request{}
		return
	}

	var url string
	var cData interface{}
	var matchHistoryArr []interface{}
	list := data["idList"].([]interface{})
	for _, matchId := range list {
		url = getMatchHistoryUrl(matchId.(string))
		cRequest, err := http.NewRequest("GET", url, nil)
		if errResponse := errorResponse(err); errResponse != blank {
			fmt.Fprint(globalRequest.getWriter(), string(errResponse.([]byte)))
			w = nil
			r = nil
			globalRequest = Request{}
			return
		}
		cData = getCurlData(cRequest)
		matchHistoryArr = append(matchHistoryArr, cData)
	}

	response := setAndGetResponse(true, "Başarılı.", matchHistoryArr, 200).([]byte)

	fmt.Fprint(globalRequest.getWriter(), string(response))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func getMatchHistorListUrl(puuId string, queue string, queueType string, offset string, limit string) string {
	var url string = "https://europe.api.riotgames.com/lol/match/v5/matches/by-puuid/" + puuId + "/ids?"

	if q := queue; q != "" {
		url += "queue=" + queue + "&"
	}
	if qT := queueType; qT != "" {
		url += "type=" + queueType + "&"
	}

	url += "start=" + offset + "&count=" + limit + "&api_key="
	return getUrlWithApiKey(url)
}

func getMatchHistoryUrl(matchId string) string {
	return getUrlWithApiKey("https://europe.api.riotgames.com/lol/match/v5/matches/" + matchId + "?api_key=")
}

func riotTxt(w http.ResponseWriter, r *http.Request) {
	dat, err := os.ReadFile("./riot.txt")
	if errResp := errorResponse(err); errResp != blank {
		fmt.Fprint(globalRequest.getWriter(), string(errResp.([]byte)))
		return
	}
	fmt.Fprint(w, string(dat))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func (r Request) getWriter() http.ResponseWriter {
	return r.writer
}

func (r Request) getRequest() *http.Request {
	return r.request
}

func (r Response) getResponse() Response {
	return r
}

func setAndGetResponse(success bool, message string, data interface{}, code int) interface{} {
	var response interface{}
	globalResponse = Response{Success: success, Message: message, Data: data, Code: code}

	successResponse, err := json.Marshal(globalResponse)
	fatalResponse := errorResponse(err)

	if response = successResponse; fatalResponse != blank {
		response, _ = json.Marshal(fatalResponse)
	}

	return response
}

func errorResponse(err error) interface{} {
	if err != nil {
		fatalResponse := setAndGetResponse(false, err.Error(), nil, http.StatusBadRequest)
		return fatalResponse
	}
	return blank
}

func tokenResponse(message string) interface{} {
	fatalResponse := setAndGetResponse(false, message, nil, http.StatusForbidden)
	return fatalResponse
}

func checkToken(token string) (errResp interface{}) {
	if len(token) > 0 {
		if db, errResp = setAndGetDb(); errResp != blank {
			return errResp
		}

		queryString := "SELECT id, token, is_active FROM settings WHERE token = ? LIMIT 1"
		row := db.QueryRow(queryString, token)

		var tokenRow TokenRow

		err := row.Scan(&tokenRow.id, &tokenRow.token, &tokenRow.isActive)
		if errResp = errorResponse(err); errResp != blank {
			return tokenResponse("Sorry, not allowed.")
		}

		if tokenRow.isActive == 0 {
			return tokenResponse("Token is expired.")
		}

		if errResp := setTokenExpired(tokenRow.id); errResp != blank {
			return errResp
		}

		return blank
	}

	return tokenResponse("Sorry, not allowed.")
}

func setTokenExpired(tokenId int) interface{} {
	queryString := "UPDATE settings SET is_active = ? WHERE id = ?"
	result, err := db.Exec(queryString, 0, tokenId)
	if errResp := errorResponse(err); errResp != blank {
		return errResp
	}

	if affectedRows, err := result.RowsAffected(); affectedRows > 0 {
		if errResp := errorResponse(err); errResp != blank {
			return errResp
		}
		return blank
	}

	return errorResponse(errors.New("Bir problem meydana geldi."))
}

func getSummonerProfileUrl(server string, userName string) string {
	summonerUrl = getUrlWithApiKey("https://" + server + ".api.riotgames.com/lol/summoner/v4/summoners/by-name/" + userName + "?api_key=")

	return summonerUrl
}

func getUrlWithApiKey(url string) string {
	return url + apiKey
}

func getCurlData(request *http.Request) interface{} {
	var data interface{}
	response, _ := http.DefaultClient.Do(request)

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	json.Unmarshal(body, &data)

	return data
}

func getDbCredentials() string {
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")

	//println(dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName)
	//db_user:password@tcp(localhost:3306)/my_db
	return dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName
}

func setAndGetDb() (db *sqlx.DB, errResp interface{}) {
	var err error
	db, err = sqlx.Connect("mysql", getDbCredentials())
	if errResp := errorResponse(err); errResp != blank {
		return db, errResp
	}

	return db, errResp
}
