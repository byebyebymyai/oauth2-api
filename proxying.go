package main

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
)

type UserServiceMiddleware func(next UserService) UserService

func makeProxyUserService(ctx context.Context, instance string) UserServiceMiddleware {
	// If instances is empty, don't proxy.
	if instance == "" {
		return func(next UserService) UserService { return next }
	}

	e1 := proxyUserAllEndpoint(ctx, instance)
	e2 := proxyUserGetEndpoint(ctx, instance)

	// And finally, return the ServiceMiddleware, implemented by proxyUserService.
	return func(next UserService) UserService {
		return proxyUserService{ctx, next, e1, e2}
	}
}

type TokenServiceMiddleware func(next oauth2.AccessGenerate) oauth2.AccessGenerate

func makeProxyTokenService(ctx context.Context, instance string) TokenServiceMiddleware {
	if instance == "" {
		return func(next oauth2.AccessGenerate) oauth2.AccessGenerate { return next }
	}

	e := proxyTokenEndpoint(ctx, instance)
	return func(next oauth2.AccessGenerate) oauth2.AccessGenerate {
		return proxyTokenService{ctx, next, e}
	}
}
