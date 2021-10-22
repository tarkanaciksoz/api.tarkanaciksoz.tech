package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Globals interface {
	getWriter() http.ResponseWriter
	getRequest() *http.Request
	getResponse() Response
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
	if dbToken, errResp := checkToken(token); errResp != nil && errResp != blank && len(dbToken) <= 0 {
		fmt.Fprint(globalRequest.getWriter(), string(errResp.([]byte)))
		return false
	}

	return true
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/getUser", getUser)
	http.HandleFunc("/riot.txt", riotTxt)
	http.ListenAndServe(":3000", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		return
	}
	var respon interface{}

	response := setAndGetResponse(true, "başarılıydı", respon, 200).([]byte)

	fmt.Fprint(globalRequest.getWriter(), string(response))
	w = nil
	r = nil
	globalRequest = Request{}
	return
}

func getUser(w http.ResponseWriter, r *http.Request) {
	if check := construct(w, r); !check {
		return
	}

	requestBody, _ := ioutil.ReadAll(globalRequest.getRequest().Body)
	if len(requestBody) <= 0 {
		response := setAndGetResponse(false, "Body is empty.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
		return
	}
	json.Unmarshal(requestBody, &requestData)
	data := requestData.(map[string]interface{})

	if (data == nil) || data["server"] == nil || data["userName"] == nil {
		response := setAndGetResponse(false, "Required values haven't given.", nil, http.StatusBadRequest).([]byte)
		fmt.Fprint(globalRequest.getWriter(), string(response))
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

func riotTxt() {
	dat, err := os.ReadFile("./riot.txt")
	if errResp := errorResponse(err); errResp != blank {
		fmt.Fprint(globalRequest.getWriter(), string(errResp.([]byte)))
		return
	}
	fmt.Print(string(dat))
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

func notAllowedError() interface{} {
	fatalResponse := setAndGetResponse(false, "Sorry not allowed.", nil, http.StatusForbidden)
	return fatalResponse
}

func checkToken(token string) (dbToken string, errResp interface{}) {
	if len(token) > 0 {
		db, err := sqlx.Connect("mysql", getDbCredentials())
		if errResp := errorResponse(err); errResp != blank {
			return "", errResp
		}

		queryString := "SELECT token FROM settings WHERE token = ? LIMIT 1"
		rows := db.QueryRow(queryString, token)
		if errResp := errorResponse(err); errResp != blank {
			return "", errResp
		}

		err = rows.Scan(&dbToken)
		if err != nil {
			errResp := notAllowedError()
			return "", errResp
		}

		if len(dbToken) > 0 {
			return dbToken, blank
		}
	}
	return "", notAllowedError()
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
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	return dbUser + ":" + dbPass + "@/" + dbName
}
