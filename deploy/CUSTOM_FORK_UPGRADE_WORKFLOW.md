# Sub2API 二开版部署与升级工作流

这份文档用于当前仓库的二开版本，目标是：

- 不直接依赖官方镜像，避免上线后二开功能被覆盖
- 尽量减少未来同步上游 `sub2api` 更新时的冲突范围
- 保持部署方式稳定、可回滚、可持续升级

---

## 1. 当前建议

不要直接使用官方镜像：

- `weishaw/sub2api:latest`

因为它不包含当前仓库里的二开改动。

请改为：

- 使用当前源码构建自己的镜像
- 使用自己的 Git 仓库保存二开代码

本仓库已新增覆盖文件：

- `deploy/docker-compose.source.yml`

它不会改写官方提供的 `deploy/docker-compose.local.yml` / `deploy/docker-compose.yml`，只是在启动时覆盖 `sub2api` 服务的镜像来源为“本地源码构建”。

---

## 2. 推荐部署方式

### 2.0 一键安装脚本

如果你想直接一键安装当前二开版本，可以在全新 Linux 服务器上执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xftk666-alt/apierkai/main/deploy/install-custom-source.sh | sudo bash
```

脚本会自动完成：

- 安装基础依赖
- 安装 Docker / Docker Compose
- 拉取当前二开仓库源码
- 生成 `deploy/.env`
- 写入默认数据库配置
- 构建并启动当前二开版本

当前默认值如下（可通过环境变量覆盖）：

- `POSTGRES_USER=xwqwert`
- `POSTGRES_DB=xwqwert`
- `POSTGRES_PASSWORD=xw123456`

例如你也可以这样自定义端口后再安装：

```bash
SERVER_PORT=18080 curl -fsSL https://raw.githubusercontent.com/xftk666-alt/apierkai/main/deploy/install-custom-source.sh | sudo bash
```

后续升级可在服务器上执行：

```bash
sudo bash /opt/sub2api/deploy/upgrade-custom-source.sh
```

### 2.1 首次部署

服务器目录示例：

```bash
/opt/sub2api
```

进入部署目录：

```bash
cd /opt/sub2api/deploy
```

复制环境变量文件：

```bash
cp .env.example .env
```

至少设置这些值：

- `POSTGRES_PASSWORD`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `JWT_SECRET`
- `TOTP_ENCRYPTION_KEY`
- `SERVER_PORT`

创建数据目录：

```bash
mkdir -p data postgres_data redis_data
```

启动二开版：

```bash
docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d --build
```

查看日志：

```bash
docker compose -f docker-compose.local.yml -f docker-compose.source.yml logs -f sub2api
```

停止服务：

```bash
docker compose -f docker-compose.local.yml -f docker-compose.source.yml down
```

---

## 3. 为什么要用覆盖文件

不要直接大改官方的 `deploy/docker-compose.local.yml`，原因是：

- 以后同步上游时更容易冲突
- 容易忘记哪些改动是官方的，哪些是你自己的
- 覆盖文件更适合长期维护

当前做法是：

- 官方文件继续保留
- 你的源码构建逻辑只放在 `deploy/docker-compose.source.yml`

以后如果上游更新了 `deploy/docker-compose.local.yml`，你仍然可以继续使用：

```bash
docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d --build
```

---

## 4. Git 仓库建议

### 4.1 不要长期直接用官方仓库当唯一远端

你现在应该准备：

- 一个自己的 Git 仓库：保存你的二开代码
- 一个上游远端 `upstream`：指向官方 `sub2api`

推荐结构：

```bash
origin   -> 你的仓库
upstream -> 官方仓库
```

### 4.2 建议命令

如果你已经 fork 了自己的仓库，推荐在本地这样整理远端：

```bash
git remote rename origin upstream
git remote add origin <你的仓库地址>
git fetch upstream
git push -u origin HEAD
```

如果你暂时还没有自己的仓库，请先创建一个，再执行上面的命令。

---

## 5. 以后如何同步上游更新

### 5.1 标准流程

先拉上游最新代码：

```bash
git fetch upstream
```

切到你的二开分支：

```bash
git checkout <你的二开分支>
```

合并上游：

```bash
git merge upstream/main
```

如果官方默认分支不是 `main`，请改成实际分支名。

处理冲突后，重新构建并上线：

```bash
cd deploy
docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d --build
```

---

## 6. 每次升级前必须做的事

升级前先备份这三个目录：

```bash
deploy/data
deploy/postgres_data
deploy/redis_data
```

建议最少做一份归档：

```bash
tar czf sub2api-backup-$(date +%F-%H%M%S).tar.gz deploy/data deploy/postgres_data deploy/redis_data
```

---

## 7. 上线建议

建议使用你自己的镜像 tag，而不是永远只用 `latest`。

例如：

```bash
SUB2API_IMAGE_TAG=v1
SUB2API_IMAGE_TAG=v2
SUB2API_IMAGE_TAG=2026-03-19
```

示例：

```bash
SUB2API_IMAGE_TAG=2026-03-19 docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d --build
```

这样做的好处：

- 更容易回滚
- 更容易知道当前线上版本
- 更适合后续多次升级

---

## 8. 回滚方式

如果新版本有问题，可以切回旧 tag 重新启动：

```bash
SUB2API_IMAGE_TAG=<旧版本标签> docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d
```

如果数据库结构已经发生变化，回滚前请确认：

- 当前版本迁移是否兼容旧版本
- 是否需要先恢复数据库备份

---

## 9. 对当前二开方案的判断

当前这套二开已经尽量按“可持续升级”思路实现：

- 支付渠道配置化，避免硬编码
- 尽量使用新增模块和设置扩展
- 尽量不直接破坏原有 iframe 购买模式
- 前后台 UI 尽量贴合原项目风格

但要明确：

- **它不是插件系统**
- **未来同步上游时仍需要做 Git 合并**

也就是说：

- 可以长期维护
- 不能完全零成本升级

这是正常的二开维护模式。
