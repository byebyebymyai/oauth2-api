package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/byebyebymyai/oauth2-api/endpoint"
	"github.com/byebyebymyai/oauth2-api/ent"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/google/uuid"
)

type UserService interface {
	All(user User) ([]User, error)
	Get(userID uuid.UUID) (User, error)
}

type defaultUserService struct {
}

func (svc defaultUserService) All(user User) ([]User, error) {
	return []User{}, nil
}

func (svc defaultUserService) Get(userID uuid.UUID) (User, error) {
	return User{}, nil
}

type proxyUserService struct {
	ctx         context.Context
	next        UserService
	allEndpoint endpoint.Endpoint
	getEndpoint endpoint.Endpoint
}

func (svc proxyUserService) All(user User) ([]User, error) {
	if res, err := svc.allEndpoint(svc.ctx, user); err != nil {
		return nil, err
	} else {
		return res.([]User), nil
	}
}

func (svc proxyUserService) Get(userID uuid.UUID) (User, error) {
	if res, err := svc.getEndpoint(svc.ctx, userID); err != nil {
		return User{}, err
	} else {
		return res.(User), nil
	}
}

type defaultTokenService struct {
}

// Token implements oauth2.AccessGenerate.
func (j defaultTokenService) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (access string, refresh string, err error) {
	return "", "", nil
}

type proxyTokenService struct {
	ctx      context.Context
	next     oauth2.AccessGenerate
	endpoint endpoint.Endpoint
}

// Token implements oauth2.AccessGenerate.
func (j proxyTokenService) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (access string, refresh string, err error) {
	response, err := j.endpoint(ctx, tokenGenerationRequest{
		Iss: data.Client.GetDomain(),
		Sub: data.UserID,
		Exp: int64(data.TokenInfo.GetAccessExpiresIn().Seconds()),
		Aud: []string{data.Request.Host},
	})

	if err != nil {
		return "", "", err
	}

	access = response.([]string)[0]
	refresh = ""

	if isGenRefresh {
		buf := bytes.NewBufferString(data.Client.GetDomain())
		buf.WriteString(data.UserID)
		buf.WriteString(strconv.FormatInt(data.CreateAt.Unix(), 10))
		refresh = base64.URLEncoding.EncodeToString([]byte(uuid.NewSHA1(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}
	return access, refresh, err
}

type ClientStorage struct {
	ctx    context.Context
	client *ent.Client
}

func (c *ClientStorage) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	client, err := c.client.Oauth2Client.Get(ctx, uuid.MustParse(id))
	if err != nil {
		return nil, err
	}
	return client, nil
}
