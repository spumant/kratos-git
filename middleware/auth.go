package middleware

import (
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/net/context"
	"kratos-gorm-git/helper"
)

func Auth() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				fmt.Println("RUN...")
				auth := tr.RequestHeader().Get("Authorization")
				if auth == "" {
					return nil, errors.New("No Auth")
				}
				userCliams, err := helper.AnalyseToken(auth)
				if err != nil {
					return nil, err
				}
				if userCliams.Identity == "" {
					return nil, errors.New("No Auth")
				}

				ctx = metadata.NewServerContext(ctx, metadata.New(map[string][]string{
					"username": []string{userCliams.Name},
					"identity": []string{userCliams.Identity},
				}))
			}
			return handler(ctx, req)
		}
	}
}
