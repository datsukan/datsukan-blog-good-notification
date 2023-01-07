package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	contentful "github.com/contentful-labs/contentful-go"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// type Lang はContentfulから取得したFieldの共通構造
type Lang struct {
	Ja string `json:"ja"`
}

var from, to, password string

func send(articleID string) error {
	if err := loadEnv(); err != nil {
		return err
	}

	articleSlug, articleTitle, err := fetchArticleAttr(articleID)
	if err != nil {
		return err
	}

	message := makeMessage(articleSlug, articleTitle)

	// メール送信を行い、レスポンスを表示
	client := sendgrid.NewSendClient(password)
	if response, err := client.Send(message); err != nil {
		log.Println(err)
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return nil
}

// loadEnv は、環境変数を読み込む。
func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("could not read environment variables")
		return err
	}

	from = os.Getenv("MAIL_FROM")
	to = os.Getenv("MAIL_TO")
	password = os.Getenv("SMTP_PASSWORD")

	return nil
}

// LoadContentfulEnv はContentful SDKの接続情報を環境変数から読み込む
func loadContentfulEnv() (string, string, error) {
	token := os.Getenv("CONTENTFUL_ACCESS_TOKEN")
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")

	if token == "" || spaceID == "" {
		m := fmt.Sprintf("environment variable not set [ token: %v, spaceID: %v ]", token, spaceID)
		fmt.Println(m)
		return "", "", errors.New(m)
	}

	return token, spaceID, nil
}

// NewContentfulSDK はContentful SDKのクライアントインスタンスを生成する
func newContentfulSDK() (*contentful.Client, *contentful.Space, error) {
	token, spaceID, err := loadContentfulEnv()
	if err != nil {
		return nil, nil, err
	}

	cma := contentful.NewCMA(token)
	space, err := cma.Spaces.Get(spaceID)

	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	return cma, space, nil
}

// fieldToString はContentfulのFieldを文字列に変換する
func fieldToString(field interface{}) (string, error) {
	byte, err := json.Marshal(field)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var body Lang
	if err := json.Unmarshal(byte, &body); err != nil {
		fmt.Println(err)
		return "", err
	}

	return body.Ja, nil
}

// articleAttr は記事情報を取得する
func articleAttr(entry *contentful.Entry) (string, string, error) {
	var slug, title string
	var err error
	for attr, field := range entry.Fields {
		switch attr {
		case "slug":
			slug, err = fieldToString(field)
			if err != nil {
				fmt.Println(err)
				return "", "", err
			}
		case "title":
			value, err := fieldToString(field)
			if err != nil {
				fmt.Println(err)
				return "", "", err
			}
			title = strings.Replace(value, "/", "／", -1)
		}
	}

	return slug, title, nil
}

func fetchArticleAttr(articleID string) (string, string, error) {
	// Contentful SDK のクライアントインスタンスを生成する
	cma, space, err := newContentfulSDK()
	if err != nil {
		return "", "", err
	}

	// Contentfulから記事情報を取得する
	entry, err := cma.Entries.Get(space.Sys.ID, articleID)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	// Contentfulから記事情報が取得できない場合、処理を終了する
	if entry == nil {
		return "", "", errors.New("article not found")
	}

	slug, title, err := articleAttr(entry)
	if err != nil {
		return "", "", err
	}

	return slug, title, nil
}

// makeMessage は、メールメッセージを構築する。
func makeMessage(articleSlug string, articleTitle string) *mail.SGMailV3 {
	message := mail.NewV3Mail()

	// 送信元を設定
	message.SetFrom(mail.NewEmail("datsukan blog", from))

	// 宛先を指定
	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", to))
	p.SetSubstitution("%articleSlug%", articleSlug)
	p.SetSubstitution("%articleTitle%", articleTitle)
	message.AddPersonalizations(p)

	// 件名を設定
	message.Subject = "記事がいいねされました！"
	// テキストパートを設定
	c := mail.NewContent("text/plain", "記事名：%articleTitle%\nhttps://blog.datsukan.me/%articleSlug%\n\n記事がいいねされました！")
	message.AddContent(c)

	return message
}
