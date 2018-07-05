package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/pat"
	yaml "gopkg.in/yaml.v2"
)

type doc struct {
	Title string  `json:"title"`
	Steps []*step `json:"steps"`

	TotalSteps     int `json:"total_steps"`
	CompletedSteps int `json:"completed_steps"`
}

type step struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Test        string `json:"test"`
	Expected    string `json:"expected"`
	Install     string `json:"install"`

	OK     bool   `json:"ok"`
	Actual string `json:"actual"`
}

var d = &doc{}

func loadDocs() {
	b, err := ioutil.ReadFile("example/doc.yml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(b, d)
	if err != nil {
		panic(err)
	}

	d.TotalSteps = len(d.Steps)

	fmt.Printf("%+v\n", d)
}

func checkState() {
	var completed int
	defer func() {
		d.CompletedSteps = completed
	}()

	for _, v := range d.Steps {
		log.Printf("Testing [%s]", v.Test)

		cmd := exec.Command("bash", "-c", v.Test)

		b, err := cmd.CombinedOutput()

		if err != nil {
			log.Printf("Error [%s]: %s", v.Test, err)
			v.Actual = err.Error()
			continue
		}

		v.Actual = string(b)

		if !strings.HasPrefix(v.Actual, v.Expected) {
			log.Printf("NOT OK [%s]: Expected [%s], Actual [%s]", v.Test, v.Expected, v.Actual)
			continue
		}

		log.Printf("OK [%s]: Expected [%s], Actual [%s]", v.Test, v.Expected, v.Actual)
		v.OK = true
		completed++
	}
}

func main() {
	loadDocs()
	checkState()

	p := pat.New()

	p.Get("/steps", func(w http.ResponseWriter, req *http.Request) {
		jsonb, err := json.Marshal(&d)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(jsonb)
	})

	if err := http.ListenAndServe(":5555", p); err != nil {
		panic(err)
	}
}
