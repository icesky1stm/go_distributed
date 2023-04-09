package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

func RegisterService(r Registration) error {

	// 注册更新服务
	serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
	if err != nil {
		return err
	}
	http.Handle(serviceUpdateURL.Path, &serviceUpdateHandler{})

	// 心跳服务
	heartBeatURL, err := url.Parse(r.HeartBeatURL)
	if err != nil {
		return err
	}
	http.HandleFunc(heartBeatURL.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		log.Println("收到心跳报文,响应成功!!!")
	})

	//发送服务注册申请
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(r)
	if err != nil {
		return err
	}

	res, err := http.Post(ServiceURL, "application/json", buf)
	if err != nil {
		return fmt.Errorf("连接注册中心失败[%v]", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("向注册中心[注册]服务失败[%v]", res.StatusCode)
	}

	return nil
}

// 更新提供者的方法
type serviceUpdateHandler struct{}

func (suh serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Update服务收到请求[%+v]\n", p)
	prov.Update(p)

}

func ShutdownService(url string) error {
	// 这段相当于手写的http请求，后续可以研究一下socket的
	req, err := http.NewRequest(http.MethodDelete, ServiceURL, bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("向注册中心[取消]服务失败[%v]", res.StatusCode)
	}

	return nil
}

// 服务提供者的处理
type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

func (p *providers) Update(pat patch) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 新增
	for _, pathEntry := range pat.Added {
		if _, ok := p.services[pathEntry.Name]; !ok {
			p.services[pathEntry.Name] = make([]string, 0)
		}

		p.services[pathEntry.Name] = append(p.services[pathEntry.Name], pathEntry.URL)
	}

	// 删除
	for _, pathEntry := range pat.Removed {
		if providerURLs, ok := p.services[pathEntry.Name]; !ok {
			for i := range providerURLs {
				if providerURLs[i] == pathEntry.URL {
					p.services[pathEntry.Name] = append(providerURLs[:i], providerURLs[i+1:]...)
				}

			}
		}

		p.services[pathEntry.Name] = append(p.services[pathEntry.Name], pathEntry.URL)
	}

}

// 没有返回string slice, 随机选取了一个URL
func (p providers) get(name ServiceName) (string, error) {
	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("没有 providers 为服务[%v]", name)
	}

	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil

}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    new(sync.RWMutex),
}
