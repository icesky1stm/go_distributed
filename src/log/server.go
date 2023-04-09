package log

// 把post请求的内容写入到log中去

import (
	"io/ioutil"
	stlog "log"
	"net/http"
	"os"
)

var log *stlog.Logger

type fileLog string

// 实现了Writer接口，因此是io.Writer类型
func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(data)
}

func Run(dest string) {
	log = stlog.New(fileLog(dest), "[go] - ", stlog.LstdFlags)
}

func RegisterHandlers() {

	// 默认注册到http.Server中作为handler
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		}
	})
}

func write(message string) {
	log.Printf("%v\n", message)
}
