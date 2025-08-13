# Register Configuration Parameters Usage Guide

This module supports setting the registration center configuration via **command-line flags** or **environment variables**.     
No need to manually add robots, the management backend can directly access them
---

## 1. Command-Line Flags

| Flag Name        | Type   | Default Value           | Description                                  |
|------------------|--------|-------------------------|----------------------------------------------|
| `-register_type` | string | `""`                    | The type of register center, e.g. `etcd`     |
| `-etcd_urls`     | string | `http://127.0.0.1:2379` | Etcd URLs, multiple URLs separated by commas |
| `-etcd_username` | string | `""`                    | Etcd username for authentication             |
| `-etcd_password` | string | `""`                    | Etcd password for authentication             |

### Example

```bash
./MuseBot-darwin-amd64 -register_type=etcd -etcd_urls="http://10.0.0.1:2379,http://10.0.0.2:2379" -etcd_username=admin -etcd_password=123456
./MuseBot-admin-darwin-amd64  -register_type=etcd -etcd_urls="http://10.0.0.1:2379,http://10.0.0.2:2379"
```

---

## 2. Environment Variables

| Environment Variable | Description                                  | Format                                               |
|----------------------|----------------------------------------------|------------------------------------------------------|
| `REGISTER_TYPE`      | Register center type, e.g. `etcd`            | Plain string                                         |
| `ETCD_URLS`          | Etcd URLs, multiple URLs separated by commas | Example: `http://10.0.0.1:2379,http://10.0.0.2:2379` |
| `ETCD_USERNAME`      | Etcd username for authentication             | Plain string                                         |
| `ETCD_PASSWORD`      | Etcd password for authentication             | Plain string                                         |

### Example (Linux/macOS)

```bash
export REGISTER_TYPE=etcd
export ETCD_URLS="http://10.0.0.1:2379,http://10.0.0.2:2379"
export ETCD_USERNAME=admin
export ETCD_PASSWORD=123456

./MuseBot-darwin-amd64 
./MuseBot-admin-darwin-amd64
```

---

## Parameter Priority

* Environment variables will override command-line flags if both are set.

---

## Notes

* `etcd_urls` or `ETCD_URLS` supports multiple Etcd node URLs separated by commas.
* Currently, supported `register_type` example is `"etcd"`; more types can be added based on your business needs.
* For security reasons, avoid exposing sensitive information such as passwords in public environment variables or
  command-line histories.

