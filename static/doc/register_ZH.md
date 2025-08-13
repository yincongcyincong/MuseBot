# 注册配置参数使用指南

本模块支持通过 **命令行参数** 或 **环境变量** 来设置注册中心的相关配置。
无需手动添加机器人，管理后台可以直接访问它们。

---

## 1. 命令行参数

| 参数名              | 类型  | 默认值                     | 说明                |
| ---------------- | --- | ----------------------- | ----------------- |
| `-register_type` | 字符串 | `""`                    | 注册中心类型，例如 `etcd`  |
| `-etcd_urls`     | 字符串 | `http://127.0.0.1:2379` | Etcd 地址，多个地址用逗号分隔 |
| `-etcd_username` | 字符串 | `""`                    | Etcd 认证用户名        |
| `-etcd_password` | 字符串 | `""`                    | Etcd 认证密码         |

### 示例

```bash
./MuseBot-darwin-amd64 -register_type=etcd -etcd_urls="http://10.0.0.1:2379,http://10.0.0.2:2379" -etcd_username=admin -etcd_password=123456
./MuseBot-admin-darwin-amd64 -register_type=etcd -etcd_urls="http://10.0.0.1:2379,http://10.0.0.2:2379"
```

---

## 2. 环境变量

| 环境变量名           | 说明                | 格式                                             |
| --------------- | ----------------- | ---------------------------------------------- |
| `REGISTER_TYPE` | 注册中心类型，例如 `etcd`  | 纯字符串                                           |
| `ETCD_URLS`     | Etcd 地址，多个地址用逗号分隔 | 示例：`http://10.0.0.1:2379,http://10.0.0.2:2379` |
| `ETCD_USERNAME` | Etcd 认证用户名        | 纯字符串                                           |
| `ETCD_PASSWORD` | Etcd 认证密码         | 纯字符串                                           |

### 示例（Linux/macOS）

```bash
export REGISTER_TYPE=etcd
export ETCD_URLS="http://10.0.0.1:2379,http://10.0.0.2:2379"
export ETCD_USERNAME=admin
export ETCD_PASSWORD=123456

./MuseBot-darwin-amd64
./MuseBot-admin-darwin-amd64
```

---

## 参数优先级

* 当同时设置了命令行参数和环境变量时，环境变量的配置会覆盖命令行参数。

---

## 注意事项

* `etcd_urls` 或 `ETCD_URLS` 支持多个 Etcd 节点地址，地址之间用英文逗号 `,` 分隔。
* 当前支持的 `register_type` 示例为 `"etcd"`，后续可根据业务需求扩展更多类型。
* 出于安全考虑，请避免在公共环境变量或命令行历史中暴露密码等敏感信息。

