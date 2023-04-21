# API Server

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![golang](https://img.shields.io/badge/golang-v1.17.13-brightgreen)](https://go.dev/doc/install)
[![version](https://img.shields.io/badge/version-v0.2.0-green)]()

Nautes 的设计是遵循了 GitOps 的最佳实践，将用户应用环境以及 Nautes 自身环境的声明均存储在版本库中。声明数据分为两类：密钥数据是存储在 Vault 中，其他数据是存储在 Git（目前只支持 GitLab）仓库中，API Server 项目则是提供了一组用于操作这些声明数据的 REST API。

## 功能简介

API Server 提供的 API 分为两类：租户侧和用户侧。

| 分类   | 功能       | 描述                                       |
| ---- | :------- | :--------------------------------------- |
| 租户侧  | 集群       | 租户管理集群注册或删除用户集群的接口，集群包括宿主集群和运行时集群，也包括物理集群和虚拟集群，使用此接口的用户需要对租户配置库有写权限。 |
| 用户侧  | 产品       | Product Provider（目前只支持 GitLab）的产品维护接口，产品在 GitLab 中唯一对应一个 Group，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |
| 用户侧  | 项目       | 产品中项目的维护接口，项目的元数据存储在 GitLab 中，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |
| 用户侧  | 代码库      | 产品中代码库的维护接口，代码库的元数据存储在 GitLab 中，代码库在 GitLab 中唯一对应一个 Project，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |
| 用户侧  | 环境       | 产品中环境的维护接口，环境的元数据存储在 GitLab 中，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |
| 用户侧  | 部署运行时    | 产品中部署运行时的维护接口，部署运行时的元数据存储在 GitLab 中，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |
| 用户侧  | 项目流水线运行时 | 项目中流水线运行时的维护接口，流水线运行时的元数据存储在 GitLab 中，使用此接口的用户只需一个 GitLab 账号，无需其他权限。 |

## API文档

如果您是在本地启动 API Server 的服务，可以通过下面的地址访问 swagger-ui 风格的 API 文档。

```shell
https://$api-server:$port/q/swagger-ui
```

您也可以通过[在线文档](https://nautes.io/api/)浏览 API 的定义。

## 快速开始

### 准备

安装以下工具，并配置 GOBIN 环境变量：

- [go](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)
- [kratos](https://go-kratos.dev/docs/getting-started/usage/#%E5%AE%89%E8%A3%85)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

准备一个 kubernetes 实例，复制 kubeconfig 文件到 {$HOME}/.kube/config

创建 nautes-configs 配置文件

```
kubectl create cm nautes-configs -n nautes
```

### 构建

```
make build
```

### 运行

```bash
kratos run
```
### 单元测试

```shell
go test -v ./...
```

