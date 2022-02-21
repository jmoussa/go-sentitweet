package monitoring

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/jmoussa/go-sentitweet/config"
)

type Log struct {
	Message   string `json:"log_message"`
	Level     string `json:"level"`
	Type      string `json:"source"`
	Timestamp string `json:"timestamp"`
}

func GetTimestamp() string {
	return fmt.Sprintf("%v", time.Now().UTC())
}

func SendLogMessageToSNS(msgPtr *Log) (sns.PublishOutput, error) {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file. (~/.aws/credentials).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sns.New(sess)
	var cfg config.Config = config.ParseConfig()
	topicArn := cfg.General["aws_topic_arn"]
	msg, err := json.Marshal(msgPtr)
	if err != nil {
		return sns.PublishOutput{}, err
	}
	msgStr := string(msg)
	topic := cfg.General["aws_logging_topic"]
	result, err := svc.Publish(&sns.PublishInput{
		Message:        &msgStr,
		MessageGroupId: &topic,
		TopicArn:       &topicArn,
	})
	if err != nil {
		fmt.Printf("*** Error publishing to SNS: %v\n", err)
		fmt.Println(err.Error())
		var emptyResult sns.PublishOutput
		return emptyResult, err
	}
	fmt.Println("Success, message id:", *result.MessageId)
	return *result, nil
}
