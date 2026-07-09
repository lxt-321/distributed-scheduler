package registry

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dscheduler/config"
)

// EtcdClient 基于 etcd v3 gRPC-Gateway(HTTP/JSON) 的轻量客户端。
// 不依赖官方 etcd Go client，避免重型依赖；直接调用 2379 端口的 HTTP 网关。
type EtcdClient struct {
	endpoint string
	http     *http.Client
}

type kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Lease int64  `json:"lease,omitempty"`
}

// NewEtcdClient 创建 etcd 客户端（取配置中第一个 endpoint）
func NewEtcdClient() *EtcdClient {
	eps := config.Global.Etcd.Endpoints
	ep := "http://127.0.0.1:2379"
	if len(eps) > 0 {
		ep = eps[0]
	}
	return &EtcdClient{
		endpoint: strings.TrimRight(ep, "/"),
		http:     &http.Client{Timeout: 5 * time.Second},
	}
}

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func (e *EtcdClient) post(path string, body interface{}) ([]byte, error) {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, e.endpoint+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return respBody, fmt.Errorf("etcd %s 返回 %d: %s", path, resp.StatusCode, respBody)
	}
	return respBody, nil
}

// GrantLease 申请租约，返回租约 ID
func (e *EtcdClient) GrantLease(ttl int64) (int64, error) {
	resp, err := e.post("/v3/lease/grant", map[string]interface{}{"TTL": ttl})
	if err != nil {
		return 0, err
	}
	var r struct {
		ID    json.RawMessage `json:"ID"`
		Error string          `json:"error"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return 0, err
	}
	if r.Error != "" {
		return 0, fmt.Errorf("etcd grant lease: %s", r.Error)
	}
	return parseLeaseID(r.ID)
}

// parseLeaseID 兼容 etcd 网关返回的 number / string 两种 JSON 形态
func parseLeaseID(raw json.RawMessage) (int64, error) {
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, "\"") {
		var id int64
		if err := json.Unmarshal(raw, &id); err != nil {
			return 0, err
		}
		return id, nil
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, err
	}
	return int64(f), nil
}

// KeepAlive 续租（单次，unary）。
// 注意：etcd gRPC-Gateway 的 keepalive 为流式接口，直接 ReadAll 会阻塞，
// 因此本实现采用「周期性重新申请租约并覆盖写入」的方式维持心跳（见 RegisterExecutor）。
func (e *EtcdClient) KeepAlive(leaseID int64) error {
	_, err := e.post("/v3/lease/keepalive", map[string]interface{}{"ID": leaseID})
	return err
}

// PutWithLease 写入带租约的 key（租约过期后 key 自动删除，实现心跳下线）
func (e *EtcdClient) PutWithLease(key, value string, leaseID int64) error {
	_, err := e.post("/v3/kv/put", map[string]interface{}{
		"key":   b64(key),
		"value": b64(value),
		"lease": leaseID,
	})
	return err
}

// Put 写入普通 key
func (e *EtcdClient) Put(key, value string) error {
	_, err := e.post("/v3/kv/put", map[string]interface{}{
		"key":   b64(key),
		"value": b64(value),
	})
	return err
}

// GetPrefix 前缀查询，返回 map[key]=value
func (e *EtcdClient) GetPrefix(prefix string) (map[string]string, error) {
	resp, err := e.post("/v3/kv/range", map[string]interface{}{
		"key":       b64(prefix),
		"range_end": b64(prefixEnd(prefix)),
	})
	if err != nil {
		return nil, err
	}
	var r struct {
		Kvs []kv `json:"kvs"`
	}
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, kvv := range r.Kvs {
		k, _ := base64.StdEncoding.DecodeString(kvv.Key)
		v, _ := base64.StdEncoding.DecodeString(kvv.Value)
		out[string(k)] = string(v)
	}
	return out, nil
}

// DeletePrefix 删除前缀
func (e *EtcdClient) DeletePrefix(prefix string) error {
	_, err := e.post("/v3/kv/delete_range", map[string]interface{}{
		"key":       b64(prefix),
		"range_end": b64(prefixEnd(prefix)),
	})
	return err
}

// prefixEnd 计算前缀的 range_end（etcd 半开区间 [key, range_end)）
func prefixEnd(prefix string) string {
	b := []byte(prefix)
	for i := len(b) - 1; i >= 0; i-- {
		b[i]++
		if b[i] != 0 {
			break
		}
	}
	return string(b)
}

// RegisterExecutor 注册执行器到 etcd 并后台维持心跳。
// key 结构： prefix + appName + "/" + address
// stop 通道关闭时自动清理注册信息（进程退出）。
func (e *EtcdClient) RegisterExecutor(prefix, appName, address string, ttl int, stop <-chan struct{}) error {
	key := prefix + appName + "/" + address
	leaseID, err := e.GrantLease(int64(ttl))
	if err != nil {
		return err
	}
	if err := e.PutWithLease(key, address, leaseID); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(time.Duration(ttl/2) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				e.DeletePrefix(key)
				return
			case <-ticker.C:
				// 重新申请租约并覆盖写入 key，等价于「心跳续租」
				// （避免使用流式 keepalive 接口导致的阻塞问题）
				if newID, gerr := e.GrantLease(int64(ttl)); gerr == nil {
					if perr := e.PutWithLease(key, address, newID); perr == nil {
						leaseID = newID
					}
				}
			}
		}
	}()
	return nil
}

// Discover 发现某 appName 下的所有在线执行器地址
func (e *EtcdClient) Discover(prefix, appName string) ([]string, error) {
	kvs, err := e.GetPrefix(prefix + appName + "/")
	if err != nil {
		return nil, err
	}
	var addrs []string
	for _, v := range kvs {
		addrs = append(addrs, v)
	}
	return addrs, nil
}
