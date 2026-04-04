package services

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMRepository interface {
	SaveToken(ctx context.Context, userID, token string) error
	GetTokenByUserID(ctx context.Context, userID string) (string, error)
}

type NotificationService interface {
	SaveFCMToken(ctx context.Context, userID, token string) error
	SendPushNotification(userID, title, body string)
}

type notificationService struct {
	repo FCMRepository
	app  *firebase.App
}

func NewNotificationService(repo FCMRepository, credsFilePath string) (NotificationService, error) {
	opt := option.WithCredentialsFile(credsFilePath)
	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "ismartshell"}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}
	return &notificationService{repo: repo, app: app}, nil
}

func (s *notificationService) SaveFCMToken(ctx context.Context, userID, token string) error {
	return s.repo.SaveToken(ctx, userID, token)
}

func (s *notificationService) SendPushNotification(userID, title, body string) {
	go func() {
		bgCtx := context.Background()
		token, err := s.repo.GetTokenByUserID(bgCtx, userID)
		if err != nil {
			log.Printf("SendPushNotification: could not get token for user %s: %v", userID, err)
			return
		}
		
		client, err := s.app.Messaging(bgCtx)
		if err != nil {
			log.Printf("SendPushNotification: messaging client error: %v", err)
			return
		}
		
		msg := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
		}
		
		response, err := client.Send(bgCtx, msg)
		if err != nil {
			log.Printf("SendPushNotification: failed to send message to user %s: %v", userID, err)
			return
		}
		
		log.Printf("SendPushNotification: successfully sent message ID %s to user %s", response, userID)
	}()
}
