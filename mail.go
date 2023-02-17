package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	attribute "github.com/datsukan/datsukan-blog-article-attribute"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var from, to, password, accessToken, spaceID string

// send は、メールを送信する。
func send(articleID string) error {
	if err := loadEnv(); err != nil {
		return err
	}

	aa, err := attribute.New(articleID, accessToken, spaceID)
	if err != nil {
		return err
	}

	if err := aa.Get(); err != nil {
		return err
	}

	message := makeMessage(aa.Slug, aa.Title)

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
	if isLocal {
		if err := godotenv.Load(); err != nil {
			fmt.Println("could not read environment variables")
			return err
		}
	}

	from = os.Getenv("MAIL_FROM")
	to = os.Getenv("MAIL_TO")
	password = os.Getenv("SMTP_PASSWORD")
	accessToken = os.Getenv("CONTENTFUL_ACCESS_TOKEN")
	spaceID = os.Getenv("CONTENTFUL_SPACE_ID")

	if accessToken == "" || spaceID == "" {
		m := fmt.Sprintf("environment variable not set [ token: %v, spaceID: %v ]", accessToken, spaceID)
		return errors.New(m)
	}

	return nil
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
