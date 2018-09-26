package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"

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
	Error  string `json:"error"`
}

var d = &doc{}

func loadDocs() {
	d = &doc{}

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
	for _, v := range d.Steps {
		checkStepState(v)
	}
}

func checkStepState(v *step) {
	var originalOK = v.OK
	defer func() {
		if originalOK && !v.OK {
			d.CompletedSteps--
		} else if !originalOK && v.OK {
			d.CompletedSteps++
		}
	}()

	log.Printf("Testing [%s]", v.Test)

	cmd := exec.Command("bash", "-c", v.Test)

	b, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Error [%s]: %s", v.Test, err)
		v.Error = err.Error()
		v.OK = false
		return
	}

	v.Actual = string(b)

	expectedRe, err := regexp.Compile(v.Expected)
	if err != nil {
		log.Printf("INVALID REGEX [%s]: %s", v.Test, v.Expected)
		v.Error = fmt.Sprintf("Invalid regex for expected output: %s", v.Expected)
		v.OK = false
		return
	}

	if !expectedRe.MatchString(v.Actual) {
		log.Printf("NOT OK [%s]: Expected [%s], Actual [%s]", v.Test, v.Expected, v.Actual)
		v.OK = false
		return
	}

	log.Printf("OK [%s]: Expected [%s], Actual [%s]", v.Test, v.Expected, v.Actual)
	v.OK = true
}

func main() {
	loadDocs()
	checkState()

	p := pat.New()

	p.Options("/reload", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	p.Post("/reload", func(w http.ResponseWriter, req *http.Request) {
		// TODO maybe make this asynchronous, and add a lock
		loadDocs()
		checkState()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	p.Get("/steps", func(w http.ResponseWriter, req *http.Request) {
		jsonb, err := json.Marshal(&d)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(jsonb)
	})

	p.Post("/steps/{id}/retest", func(w http.ResponseWriter, req *http.Request) {
		stepID := req.URL.Query().Get(":id")
		nID, err := strconv.Atoi(stepID)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		if nID < 0 || nID > len(d.Steps)-1 {
			w.WriteHeader(404)
			return
		}

		checkStepState(d.Steps[nID])
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	if err := http.ListenAndServe(":5555", p); err != nil {
		panic(err)
	}
}
