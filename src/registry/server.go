package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// 注册中心本身的配置
const ServerPort = ":3000"
const ServiceURL = "http://127.0.0.1" + ServerPort + "/services"

/*** 注册中心的所有注册内容 ***/
type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
	// 保证线程安全
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

/** 对注册中心的增删改查: 1.增加注册 **/
func (r *registry) add(reg Registration) error {
	// 加锁
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()

	// 通知自己 依赖服务列表
	err := r.sendRequiredServices(reg)
	if err != nil {
		return err
	}

	// 通知依赖服务 自己的情况
	r.notify(patch{
		Added: []pathEntry{
			pathEntry{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})

	return nil
}

func (r *registry) heartBeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup

		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()
				res, err := http.Get(reg.HeartBeatURL)
				if err != nil {
					log.Println(err)
				} else if res.StatusCode == http.StatusOK {
					log.Printf("心跳检查通过[%s][%s]\n", reg.ServiceName, reg.ServiceURL)
					return
				}
				log.Printf("心跳检查失败!!![%s][%s]\n", reg.ServiceName, reg.ServiceURL)
				// 删除服务
				r.remove(reg.ServiceURL)
				// time.Sleep(1 * time.Second)

			}(reg)
			wg.Wait()
			time.Sleep(freq)
		}

	}
}

// 只会运行一次
var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartBeat(3 * time.Second)
	})
}

func (r registry) notify(fullPatch patch) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []pathEntry{}, Removed: []pathEntry{}}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					log.Printf("需要向[%s]更新: [%v][%+v]\n", reg.ServiceUpdateURL, sendUpdate, p)
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}

}

func (r registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	// 遍历整个注册中心
	for _, serviceReg := range r.registrations {
		// 遍历当前服务的所有依赖服务
		for _, reqService := range reg.RequiredServices {
			// 如果注册中心中的服务 == 如果当前服务的依赖服务
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, pathEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

func (r registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("发送更新数据异常[%v]", res.StatusCode)
	}

	return nil
}

/** 对注册中心的增删改查: 2.删除注册 **/
func (r *registry) remove(url string) error {
	for i := range reg.registrations {
		// URL相等
		if reg.registrations[i].ServiceURL == url {
			// 通知依赖服务 自己的情况
			r.notify(patch{
				Removed: []pathEntry{
					pathEntry{
						Name: reg.registrations[i].ServiceName,
						URL:  reg.registrations[i].ServiceURL,
					},
				},
			})

			r.mutex.Lock()
			defer r.mutex.Unlock()
			// 一点也不优雅，这写法
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			return nil
		}

	}
	return nil
}

/** 接收信息主方法 **/
type RegistryService struct{}

// 实现ServeHTTP接口, 才能被外层的http.handle使用
func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received!!!")

	// 判断接收方法
	switch r.Method {
	case http.MethodPost:
		// 获取json内容
		dec := json.NewDecoder(r.Body)
		// 解析成为一个Registration
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// 打印
		log.Printf("Adding service : %v with URl: %s\n", r.ServiceName, r.ServiceURL)
		// 调用增加函数
		err = reg.add(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s ", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		// 将范围参数设置为失败
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}
