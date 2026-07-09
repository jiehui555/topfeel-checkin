# Topfeel 签到工具开发指南

## 项目概述
Go语言编写的BBS Topfeel每日签到工具，支持定时任务和单次执行。

## 环境要求
- Go 1.22+
- Docker（用于容器化部署）
- 环境变量：`TOPFEEL_TOKEN`（多个token用逗号分隔）

## 构建与运行

### 本地构建
```bash
go build -o topfeel-checkin .
```

### 运行方式
```bash
# 单次执行
./topfeel-checkin -once

# 定时任务（每天08:00执行）
./topfeel-checkin
```

### Docker构建
```bash
docker build -t topfeel-checkin .
docker run -e TOPFEEL_TOKEN="your_token" topfeel-checkin
```

## 关键注意事项

### 时区配置
- 容器必须安装`tzdata`包以支持`Asia/Shanghai`时区
- Dockerfile已包含：`RUN apk add --no-cache tzdata`
- 如果缺少时区数据，cron任务会回退到本地时间

### 环境变量
- `.env`文件用于本地开发，已被.gitignore排除
- 生产环境通过Docker环境变量注入

## CI/CD
- GitHub Actions自动构建Docker镜像
- 推送到GHCR（GitHub Container Registry）
- 触发条件：main分支推送或版本标签（v*）
- 镜像标签：latest、sha、版本号

## 代码规范
- 使用中文错误信息
- 遵循Go标准代码风格
- 错误处理要详细，包含上下文信息

## 常见问题
1. **时区错误**：确保Docker镜像包含tzdata包
2. **签到失败**：检查TOPFEEL_TOKEN是否有效
3. **网络问题**：确认能访问bbs.topfeel.com