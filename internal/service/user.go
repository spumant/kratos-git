package service

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/metadata"
	"kratos-gorm-git/helper"
	"kratos-gorm-git/models"

	pb "kratos-gorm-git/api/git"
)

type UserService struct {
	pb.UnimplementedUserServer
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	var username, identity string
	if md, ok := metadata.FromServerContext(ctx); ok {
		username = md.Get("username")
		identity = md.Get("identity")
		fmt.Println(username, identity)
	}
	ub := new(models.UserBasic)
	err := models.DB.Where("username = ? AND password = ?", req.Username, helper.GetMd5(req.Password)).Find(ub).Error
	if err != nil {
		return nil, err
	}

	token, err := helper.GenerateToken(ub.Identity, ub.Username)
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{
		Token: token,
	}, nil
}
