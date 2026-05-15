package articleauth

import (
	"context"

	ssov1 "github.com/los3-smurf/photos/gen/go/sso"
)

type ArticleAuth struct {
	client ssov1.AuthClient
	appId  int32
}

func NewArticleAuth(c ssov1.AuthClient, appId int32) *ArticleAuth {
	return &ArticleAuth{client: c, appId: appId}
}

func (a *ArticleAuth) Login(ctx context.Context, email, password string) (string, error) {
	resp, err := a.client.Login(ctx, &ssov1.LoginRequest{
		Email: email, Password: password, AppId: a.appId,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

func (a *ArticleAuth) Register(ctx context.Context, email, password string) (int64, error) {
	resp, err := a.client.Register(ctx, &ssov1.RegisterRequest{
		Email: email, Password: password,
	})
	if err != nil {
		return 0, err
	}
	return resp.UserId, nil
}
