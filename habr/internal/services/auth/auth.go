package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrEmailNotConfirmed = errors.New("email not confirmed")
)

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte, uuid string) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appId int) (models.App, error)
}

type NotificationSender interface {
	Send(ctx context.Context, email string, uuid string) error
}

type Auth struct {
	log                *slog.Logger
	usrSaver           UserSaver
	usrProvider        UserProvider
	appProvider        AppProvider
	notificationSender NotificationSender
	tokenTTL           time.Duration
}

func NewAuth(log *slog.Logger,
	usrSaver UserSaver,
	usrProvider UserProvider,
	appProvider AppProvider,
	notificationSender NotificationSender,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{log: log,
		usrSaver:           usrSaver,
		usrProvider:        usrProvider,
		appProvider:        appProvider,
		notificationSender: notificationSender,
		tokenTTL:           tokenTTL,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "Auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("register new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)

	}
	uuid := uuid.New().String()
	id, err := a.usrSaver.SaveUser(ctx, email, passHash, uuid)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = a.notificationSender.Send(ctx, email, uuid)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("try to logging user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		log.Error("failed to fetch user data", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !user.IsConfirmed {
		return "", fmt.Errorf("%s: %w", op, ErrEmailNotConfirmed)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredential)
	}
	app, err := a.appProvider.App(ctx, appID)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)

	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
