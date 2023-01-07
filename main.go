package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Input struct {
	ID string
}

func main() {
	t := flag.Bool("local", false, "ローカル実行か否か")
	ID := flag.String("id", "", "ローカル実行用の記事ID")
	flag.Parse()

	isLocal, err := isLocal(t, ID)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if isLocal {
		fmt.Println("local")
		localController(ID)
		os.Exit(0)
	}

	fmt.Println("production")
	lambda.Start(controller)
}

// controller は、AWS Lambda 上での実行処理を行う
func controller(ctx context.Context, sqsEvent events.SQSEvent) error {
	articleID, err := articleID(sqsEvent)
	if err != nil {
		return err
	}

	if err := useCase(articleID); err != nil {
		return err
	}

	return nil
}

// isLocal はローカル環境の実行であるかを判定する
func isLocal(t *bool, ID *string) (bool, error) {
	if !*t {
		return false, nil
	}

	if *ID == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だがID指定が無いので処理不可能")
	}

	return true, nil
}

// localController はローカル環境での実行処理を行う
func localController(ID *string) {
	if err := useCase(*ID); err != nil {
		fmt.Println(err.Error())
	}
}

// useCase はアプリケーションのIFに依存しないメインの処理を行う
func useCase(articleID string) error {
	return send(articleID)
}

// articleID は、SQSのイベント情報から記事IDを取得する
func articleID(sqsEvent events.SQSEvent) (string, error) {
	if len(sqsEvent.Records) == 0 {
		return "", errors.New("request content does not exist")
	}

	b := []byte(sqsEvent.Records[0].Body)
	var input Input
	if err := json.Unmarshal(b, &input); err != nil {
		return "", err
	}

	return input.ID, nil
}
