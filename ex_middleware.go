package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Course struct {
	Id         int     `json: "id"`
	Name       string  `json: "name"`
	Price      float64 `json: "price"`
	Instructor string  `json: "instructor"`
}

var CourseList []Course

func init() {
	CourseJSON := `[
		{
			"id": 1,
			"name": "Python",
			"price": 2500,
			"instructor": "BorntoDev"
		},
		{
			"id": 2,
			"name": "SQL",
			"price": 50,
			"instructor": "BorntoDev"
		}
	]`

	err := json.Unmarshal([]byte(CourseJSON), &CourseList) //json->object
	if err != nil {
		log.Fatal(err)
	}
}

func getNextId() int {
	highestId := -1
	for _, course := range CourseList {
		if highestId < course.Id {
			highestId = course.Id
		}
	}
	return highestId + 1
}

func findId(id int) (*Course, int) {
	for i, course := range CourseList {
		if course.Id == id {
			return &course, i
		}
	}
	return nil, 0
}

func courseHandler(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, "course/")
	id, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	course, index := findId(id)
	if course == nil {
		http.Error(w, fmt.Sprintf("no course with id %d", id), http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		courseJSON, err := json.Marshal(course) //object->json
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "application/json")
		w.Write(courseJSON)
	case http.MethodPut:
		var updateCourse Course
		byteBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(byteBody, &updateCourse)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if updateCourse.Id != id {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		course = &updateCourse
		CourseList[index] = *course
		updated, _ := json.Marshal(course)
		w.Header().Set("Content-type", "application/json")
		w.Write(updated)
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func coursesHandler(w http.ResponseWriter, r *http.Request) {
	courseJSON, err := json.Marshal(CourseList)
	switch r.Method {
	case http.MethodGet:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(courseJSON)
	case http.MethodPost:
		var newCourse Course
		Bodybyte, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(Bodybyte, &newCourse)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newCourse.Id = getNextId()
		CourseList = append(CourseList, newCourse)
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func middlewareHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("before handler middleware")
		handler.ServeHTTP(w, r)
		fmt.Println("middleware finished")
	})
}

func main() {
	// fmt.Println("Hello go")

	courseItem := http.HandlerFunc(courseHandler)
	courseList := http.HandlerFunc(coursesHandler)

	http.Handle("/course/", middlewareHandler(courseItem))
	http.Handle("/course", middlewareHandler(courseList))
	http.ListenAndServe(":8000", nil)
}
