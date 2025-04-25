package geecache

// 提供被其他节点访问的能力

import (
	"fmt"
	"geecache/consistenthash"
	pb "geecache/geecachepb"
	"io"

	"google.golang.org/protobuf/proto"

	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self        string                 // this peer's base URL, e.g.
	basePath    string                 // 项目名。可能会有多个项目，所以加以区分。
	mu          sync.Mutex             // guards peers and httpGetters
	peers       *consistenthash.Map    // hash环，用于记录所有的server。
	httpGetters map[string]*httpGetter // string is key of cacheserver like e.g. "http://10.0.0.2:8008"。httpgetter非常简单：一个url+一个get方法。
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { // r.URL.Path: "/_geecache/scores/Tom"
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// r.URL.Path: /_geecache/scores/Tom
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 { //做了一个简单的判错
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set updates the pool's list of peers.
// Set不仅将所有的cache server地址注册进了hash环，还将这些地址包进了httpgetter。后面的使用就是：先找hash环，再根据hash环所得
// key对应的httpgetter的Get方法来实现http请求。
func (p *HTTPPool) Set(peers ...string) { // peers[0] == "http://localhost:8001"
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil) // 只会在开始的时候调用一次
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers)) // httppool的getter和地址是分开实现的。
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) // 用来检验是否HTTPPool已经实现了接口PeerPicker

// 可以理解为http客户端，用来发出http请求的。
type httpGetter struct {
	baseURL string // e.g. "http://localhost:8001/_geecache/"
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()), // 将特殊字符进行转义。
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
