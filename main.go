package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	client := GetClient(context.Background())
	CreateTaskInCalendar(client)
}

func NewGoogleAuthConf() *oauth2.Config {

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	credentialsJSON := os.Getenv("GOOGLE_OAUTH_CREDENTIALS_JSON")
	// 第2引数に認証を求めるスコープを設定します.
	// 今回はスプレッドシートのリード権限スコープを指定.
	// config, err := google.ConfigFromJSON([]byte(credentialsJSON), sheets.SpreadsheetsReadonlyScope)
	config, err := google.ConfigFromJSON([]byte(credentialsJSON), calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

var GlobalState string

func Auth() (output string, err error) {
	conf := NewGoogleAuthConf()
	state, err := MakeRandomStr(10)
	if err != nil {
		fmt.Println(err)
		return output, err
	}

	GlobalState = state
	// stateをsessionなどに保存.
	// リダイレクトURL作成.
	redirectURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	// redirectURLをクライアントに返す.

	output = redirectURL
	fmt.Println(output)
	return output, nil
}

//state認証用の乱数生成.
func MakeRandomStr(strRange uint32) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, strRange)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error...")
	}

	// letters からランダムに取り出して文字列を生成
	var result string
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}

func Link(ctx context.Context) error {
	// クライアントからcode, stateを取得.
	// https://{call_back_uri}?state={state}&code={code}&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fspreadsheets.readonly
	code := ``
	state := `state`
	// stateが正しいか検証.

	if state != GlobalState {
		// return fmt.Errorf("stateが一致しません")
	}

	conf := NewGoogleAuthConf()
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	// CreateTaskInCalendar()
	//  token.AccessToken, token.RefreshToken, token.ExpiryをDBに保存.
	fmt.Println(token.AccessToken)
	fmt.Println(token.RefreshToken)
	fmt.Println(token.Expiry)
	return nil
}

func GetClient(ctx context.Context) *http.Client {
	conf := NewGoogleAuthConf()

	// DBに保存したトークン情報取得.
	accessToken := ``
	refreshToken := ``
	expiry, err := httpdate.Str2Time(`2022-10-05 15:47:05.713912 +0900`, nil)
	if err != nil {
		log.Fatalf("Unable to parse token expiry from string: %v", err)
	}

	token := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}

	return conf.Client(ctx, token)
}

func SpreadsheetSheetGet() error {
	// クライアントから取得したいスプレッドシートIDを受け取る.
	spreadsheetID := `1c3PCEvZn81GI5HpaqVedvVtDaYFj1AiCHtgC8QJ6Lpg`
	readRange := `A1:B2`
	// クライアント取得.
	ctx := context.Background()
	client := NewClient(ctx)
	// シート情報取得.
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}
	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get .
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return err
	}
	// クライアントにスプレッドシート情報を渡す.
	fmt.Println(resp)
	return nil
}

func CreateTaskInCalendar(client *http.Client) error {
	ctx := context.Background()
	// client := NewClient(ctx)
	fmt.Println(ctx)
	fmt.Println(client)

	// accessToken := ``
	// refreshToken := ``
	// expiry := time.Now().Add(time.Hour * 24 * 30)

	// token := &oauth2.Token{
	// 	AccessToken:  accessToken,
	// 	TokenType:    "bearer",
	// 	RefreshToken: refreshToken,
	// 	Expiry:       expiry,
	// }
	// client := oauth2.NewClient(ctx, token.AccessToken)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Unable to retrieve Calendar client: %v", err)
	}
	// Refer to the Go quickstart on how to setup the environment:
	// https://developers.google.com/calendar/quickstart/go
	// Change the scope to calendar.CalendarScope and delete any stored credentials.
	fmt.Println(time.Now().Format(time.RFC3339Nano))
	trimedTime := strings.Split(time.Now().Format(time.RFC3339Nano), "+")[0]
	fmt.Println(trimedTime)

	fmt.Println(time.Now().Add(time.Hour * 1).Format(time.RFC3339Nano))

	event := &calendar.Event{
		Summary:     "山田さん面接対策",
		Location:    "",
		Description: "株式会社テスト最終面接に向けた面接対策です。",
		Start: &calendar.EventDateTime{
			DateTime: strings.Split(time.Now().Format(time.RFC3339Nano), "+")[0],
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			DateTime: strings.Split(time.Now().Add(time.Hour*1).Format(time.RFC3339Nano), "+")[0],
			TimeZone: "Asia/Tokyo",
		},
		// Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
		// Attendees: []*calendar.EventAttendee{
		// 	&calendar.EventAttendee{Email: "lpage@example.com"},
		// 	&calendar.EventAttendee{Email: "sbrin@example.com"},
		// },
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		fmt.Printf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
	return nil
}

func GetTaskFromCalendar() {
	ctx := context.Background()
	client := NewClient(ctx)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	timeNow := time.Now().Format(time.RFC3339)
	trimedTime := strings.SplitAfterN(timeNow, "", 5)

	minTime := fmt.Sprintf(`%v-01T00:00:00`, trimedTime[0])

	//翌月1日の0時0分0秒を取得
	trimedTimeForMax := strings.SplitAfterN(trimedTime[0], "", 4)
	trimedMonth, err := strconv.Atoi(trimedTimeForMax[1])
	if err != nil {
		fmt.Println(err)
	}

	maxTime := fmt.Sprintf(`%v%vT00:00:00`, trimedTime[0], trimedMonth+1)

	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(minTime).TimeMax(maxTime).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}

func NewClient(ctx context.Context) *http.Client {
	conf := NewGoogleAuthConf()
	accessToken := ``
	refreshToken := ``
	expiry, err := httpdate.Str2Time(`2022-10-05 15:47:05.713912`, nil)
	if err != nil {
		log.Fatalf("Unable to parse token expiry from string: %v", err)
	}

	token := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
	// token取得.
	tokenSource := conf.TokenSource(ctx, token)
	// token更新.
	mySrc := &MyTokenSource{
		src:  tokenSource,
		f:    TokenRefresh,
		dbID: 1,
		// dbID: `更新するDBのレコードID`,
	}
	reuseSrc := oauth2.ReuseTokenSource(token, mySrc)
	client := oauth2.NewClient(ctx, reuseSrc)
	return client
}

type MyFunc func(*oauth2.Token, uint) error

func TokenRefresh(t *oauth2.Token, dbID uint) error {
	// 更新されたtoken情報と対象のDBレコードIDをもとにDBのToken情報を更新.
	return nil
}

type MyTokenSource struct {
	src  oauth2.TokenSource
	f    MyFunc
	dbID uint
}

func (s *MyTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	if err = s.f(t, s.dbID); err != nil {
		return t, err
	}
	return t, nil
}
