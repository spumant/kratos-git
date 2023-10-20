# Kratos gorm git

## 安装
1. 安装 cli 工具
```shell
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest 
```

2. 初始项目
```shell 
kratos new kratos-gorm-git
```

3. 运行
```shell
kratos run
```

## 相关命令
```shell
# 用户模块
# 创建 user.proto
kratos proto add api/git/user.proto
# 创建 PB
kratos proto client api/git/user.proto
# 生成 Service
kratos proto server api/git/user.proto t internal/service

# 仓库模块
# 创建 repo.proto
kratos proto add api/git/repo.proto
# 创建 PB
kratos proto client api/git/repo.proto
# 生成 Service
kratos proto server api/git/repo.proto t internal/service

# config init pb
kratos proto client internal/conf/conf.proto
```

## 核心扩展

```shell
go get github.com/asim/git-http-backend
```

## Git 远程仓库的其他实现
1. 使用git守护进程
```shell
git daemon --export-all --verbose --base-path=. --export-all --port=9091 --enable=receive-pack 
```

2. 使用http-backend
+ demo地址： git clone http://127.0.0.1:8000/git/up-zero/up-git.git
