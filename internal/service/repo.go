package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/metadata"
	"gorm.io/gorm"
	"kratos-gorm-git/define"
	"kratos-gorm-git/helper"
	"kratos-gorm-git/models"
	"os"
	"os/exec"

	pb "kratos-gorm-git/api/git"
)

type RepoService struct {
	pb.UnimplementedRepoServer
}

func NewRepoService() *RepoService {
	return &RepoService{}
}

func (s *RepoService) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoReply, error) {
	// 获取用户的基础信息
	md, exit := metadata.FromServerContext(ctx)
	if !exit {
		return nil, errors.New("no auth")
	}
	userIdentity := md.Get("identity")
	fmt.Println(userIdentity)
	ub := new(models.UserBasic)
	err := models.DB.Model(new(models.UserBasic)).Where("identity = ?", userIdentity).First(ub).Error
	if err != nil {
		return nil, err
	}
	// 查重
	var cnt int64
	err2 := models.DB.Model(new(models.RepoBasic)).Where("path = ?", req.Path).Count(&cnt).Error
	if err2 != nil {
		return nil, err2
	}
	if cnt > 0 {
		return nil, errors.New("路径已存在")
	}
	// 落库
	rb := models.RepoBasic{
		Identity: helper.GetUUID(),
		Path:     req.Path,
		Name:     req.Name,
		Desc:     req.Desc,
		Type:     int(req.Type),
	}
	ru := models.RepoUser{
		Uid:  ub.ID,
		Type: 1,
	}

	err2 = models.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(rb).Error
		if err != nil {
			return err
		}
		ru.Rid = rb.ID
		err = tx.Create(ru).Error
		if err != nil {
			return err
		}
		// init repo path
		// mkdir path
		gitRepoPath := define.RepoPath + string(os.PathSeparator) + req.Path
		err = os.Mkdir(gitRepoPath, 0755)
		if err2 != nil {
			return err
		}

		// git init --bare
		cmd := exec.Command("/bin/bash", "-c", "cd "+gitRepoPath+"; git init --bare")
		err = cmd.Run()
		if err != nil {
			return err
		}
		return nil
	})
	return &pb.CreateRepoReply{}, nil
}
func (s *RepoService) UpdateRepo(ctx context.Context, req *pb.UpdateRepoRequest) (*pb.UpdateRepoReply, error) {
	err := models.DB.Model(new(models.RepoBasic)).Where("identity = ?", req.Identity).Updates(map[string]any{
		"name": req.Name,
		"desc": req.Desc,
		"type": req.Type,
	}).Error
	if err != nil {
		return nil, err
	}
	return &pb.UpdateRepoReply{}, nil
}
func (s *RepoService) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoReply, error) {
	// 获取仓库的基础信息
	var rb = new(models.RepoBasic)
	err := models.DB.Model(new(models.RepoBasic)).Where("identity = ?", req.Identity).Find(rb).Error
	if err != nil {
		return nil, err
	}
	// 删除记录
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 删除仓库数据
		err := os.RemoveAll(define.RepoPath + string(os.PathSeparator) + rb.Path)
		if err != nil {
			return err
		}
		// 删除DB记录
		err = tx.Where("identity = ?", req.Identity).Delete(new(models.RepoBasic)).Error
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.DeleteRepoReply{}, nil
}
func (s *RepoService) GetRepo(ctx context.Context, req *pb.GetRepoRequest) (*pb.GetRepoReply, error) {
	return &pb.GetRepoReply{}, nil
}
func (s *RepoService) ListRepo(ctx context.Context, req *pb.ListRepoRequest) (*pb.ListRepoReply, error) {
	rb := make([]*models.RepoBasic, 0)
	var cnt int64
	err := models.DB.Model(new(models.RepoBasic)).Count(&cnt).Offset(int((req.Page - 1) * req.Size)).Limit(int(req.Size)).
		Find(&rb).Error
	if err != nil {
		return nil, err
	}
	list := make([]*pb.ListRepoItem, 0, len(rb))
	for _, v := range rb {
		list = append(list, &pb.ListRepoItem{
			Identity: v.Identity,
			Name:     v.Name,
			Desc:     v.Desc,
			Path:     v.Path,
			Star:     v.Star,
		})
	}
	return &pb.ListRepoReply{
		List: list,
		Cnt:  cnt,
	}, nil
}

func (s *RepoService) RepoAuth(ctx context.Context, req *pb.RepoAuthRequest) (*pb.RepoAuthReply, error) {
	// 获取用户的基础信息
	md, exit := metadata.FromServerContext(ctx)
	if !exit {
		return nil, errors.New("no auth")
	}
	userIdentity := md.Get("identity")
	ub := new(models.UserBasic)
	err := models.DB.Model(new(models.UserBasic)).Where("identity = ?", userIdentity).First(ub).Error
	if err != nil {
		return nil, err
	}

	// 获取被授权的用户的信息
	ubAuth := new(models.UserBasic)
	err = models.DB.Model(new(models.UserBasic)).Where("identity = ?", req.UserIdentity).First(ubAuth).Error
	if err != nil {
		return nil, err
	}

	// 获取仓库的基础信息
	rb := new(models.RepoBasic)
	err = models.DB.Model(new(models.RepoBasic)).Where("identity = ?", req.RepoIdentity).First(&rb).Error
	if err != nil {
		return nil, err
	}

	// 判断当前用户的权限
	var cnt int64
	err = models.DB.Model(new(models.RepoUser)).Where("rid = ? AND uid = ? AND type = 1", rb.ID, ub.ID).Count(&cnt).Error
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, errors.New("非法操作")
	}

	// 判断是否已有权限
	err = models.DB.Model(new(models.RepoUser)).Where("rid = ? AND = uid = ?", rb.ID, ubAuth.ID).Count(&cnt).Error
	if err != nil {
		return nil, err
	}
	if cnt > 0 {
		return &pb.RepoAuthReply{}, nil
	}

	// 入库
	models.DB.Create(&models.RepoUser{
		Rid:  rb.ID,
		Uid:  ubAuth.ID,
		Type: 2,
	})

	return &pb.RepoAuthReply{}, nil
}
