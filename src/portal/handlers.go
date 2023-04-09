package portal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_distributed/src/grades"
	"go_distributed/src/registry"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	http.Handle("/", http.RedirectHandler("students", http.StatusPermanentRedirect))

	h := new(studentsHandler)
	http.Handle("/students", h)
	http.Handle("/students/", h)

}

type studentsHandler struct{}

func (sh studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	switch len(pathSegments) {
	case 2: // /students
		sh.renderStudents(w, r)
	case 3: // /students/{:id}
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sh.renderStudent(w, r, id)

	case 4: // /students/{:id}/grades
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strings.ToLower(pathSegments[3]) != "grades" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sh.renderGrades(w, r, id)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (studentsHandler) renderStudents(w http.ResponseWriter, r *http.Request) {
	var err error
	// 通用返回异常
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error retrieving students: ", err)
		}

	}()

	serviceURL, err := registry.GetProvider(registry.GradingService)

	res, err := http.Get(serviceURL + "/students")
	if err != nil {
		return
	}

	var s grades.Students
	err = json.NewDecoder(res.Body).Decode(&s)
	if err != nil {
		return
	}

	rootTemplate.Lookup("students.html").Execute(w, s)

}

func (studentsHandler) renderStudent(w http.ResponseWriter, r *http.Request, id int) {
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("renderStudent 发生错误: ", err)
		}
	}()
	fmt.Printf("SSS-收到请求报文id[%d],body[%v]\n", id, r.URL.Path)

	serviceURL, err := registry.GetProvider(registry.GradingService)
	if err != nil {
		return
	}

	res, err := http.Get(fmt.Sprintf("%v/students/%v", serviceURL, id))
	if err != nil {
		return
	}

	var s grades.Student
	err = json.NewDecoder(res.Body).Decode(&s)
	if err != nil {
		return
	}

	rootTemplate.Lookup("student.html").Execute(w, s)
}

func (studentsHandler) renderGrades(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer func() {
		w.Header().Add("location", fmt.Sprintf("/students/%v", id))
		w.WriteHeader(http.StatusTemporaryRedirect)
	}()

	title := r.FormValue("Title")
	gradeType := r.FormValue("Type")
	score, err := strconv.ParseFloat(r.FormValue("Score"), 32)
	if err != nil {
		log.Println("分数 parse错误 ", err)
		return
	}

	g := grades.Grade{
		Title: title,
		Type:  grades.GradeType(gradeType),
		Score: float32(score),
	}

	data, err := json.Marshal(g)
	if err != nil {
		log.Println("转换成json异常: ", g, err)
	}

	serverURL, err := registry.GetProvider(registry.GradingService)
	if err != nil {
		return
	}

	res, err := http.Post(fmt.Sprintf("%v/students/%v/grades", serverURL, id), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Println("调用后台graedes失败 ", err)
		return
	}

	if res.StatusCode != http.StatusCreated {
		log.Println("调用后台graedes异常 ", res.StatusCode)
		return
	}
}
