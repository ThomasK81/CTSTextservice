package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Index struct {
	Urn string `json:"urn"`
}

type Node struct {
	ID   string `json:"ID"`
	Text string `json:"Text"`
}

type ServerConfig struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Source string `json:"cex_source"`
}

type CTSParams struct {
	Sourcetext, Filter string
}

func LoadConfiguration(file string) ServerConfig {
	var config ServerConfig
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func ReturnURNS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourcetext := strings.Join([]string{vars["source"], "cex"}, ".")
	result := ParseURNS(CTSParams{Sourcetext: sourcetext})
	fmt.Fprintln(w, result)
}

func ReturnSpecURNS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
  filter := vars["filter"]
	sourcetext := strings.Join([]string{vars["source"], "cex"}, ".")
	result := ParseURNS(CTSParams{Sourcetext: sourcetext, Filter: filter})
	fmt.Fprintln(w, result)
}

func ReturnNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourcetext := strings.Join([]string{vars["source"], "cex"}, ".")
	result := ParseNodes(CTSParams{Sourcetext: sourcetext})
	fmt.Fprintln(w, result)
}

func ReturnSpecNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
  filter := vars["filter"]
	sourcetext := strings.Join([]string{vars["source"], "cex"}, ".")
	result := ParseNodes(CTSParams{Sourcetext: sourcetext, Filter: filter})
	fmt.Fprintln(w, result)
}

func ParseURNS(p CTSParams) string {

	confvar := LoadConfiguration("config.json")

	input_file := confvar.Source + p.Sourcetext

	data, err := getContent(input_file)
	if err != nil {
		return "I felt a great disturbance in the Force, as if millions of requests suddenly cried out in terror and were suddenly silenced."
	}

	str := string(data)
	str = strings.Split(str, "#!ctsdata")[1]
	str = strings.Split(str, "#!")[0]

	reader := csv.NewReader(strings.NewReader(str))
	reader.Comma = '#'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = 2

	var index []Index

  switch {
  case p.Filter != "":
    for {
  		line, error := reader.Read()
  		if error == io.EOF {
  			break
  		} else if error != nil {
  			log.Fatal(error)
  		}
      if strings.Contains(line[0], p.Filter) {
        index = append(index, Index{
    			Urn: line[0],
    		})
    }
  	}
  default:
    for {
  		line, error := reader.Read()
  		if error == io.EOF {
  			break
  		} else if error != nil {
  			log.Fatal(error)
  		}
      index = append(index, Index{
  			Urn: line[0],
  		})
  	}
  }
	indexJson, _ := json.Marshal(index)
  return string(indexJson)
}

func ParseNodes(p CTSParams) string {
	confvar := LoadConfiguration("config.json")

	input_file := confvar.Source + p.Sourcetext

	data, err := getContent(input_file)
	if err != nil {
		return "I felt a great disturbance in the Force, as if millions of requests suddenly cried out in terror and were suddenly silenced."
	}

	str := string(data)
	str = strings.Split(str, "#!ctsdata")[1]
	str = strings.Split(str, "#!")[0]

	reader := csv.NewReader(strings.NewReader(str))
	reader.Comma = '#'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = 2

	var index []Node

  switch {
  case p.Filter != "":
    for {
  		line, error := reader.Read()
  		if error == io.EOF {
  			break
  		} else if error != nil {
  			log.Fatal(error)
  		}
      if strings.Contains(line[0], p.Filter) {
        index = append(index, Node{
    			ID:   line[0],
    			Text: line[1],
  		})
    }
  	}
  default:
    for {
  		line, error := reader.Read()
  		if error == io.EOF {
  			break
  		} else if error != nil {
  			log.Fatal(error)
  		}
  		index = append(index, Node{
  			ID:   line[0],
  			Text: line[1],
  		})
  	}
  }
  indexJson, _ := json.Marshal(index)
  return string(indexJson)
}

func main() {
	confvar := LoadConfiguration("./config.json")
	serverIP := confvar.Port
	router := mux.NewRouter().StrictSlash(true)
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static/").Handler(s)
	router.HandleFunc("/cex/{source}/urns", ReturnURNS)
  router.HandleFunc("/cex/{source}/urns/{filter}", ReturnSpecURNS)
	router.HandleFunc("/cex/{source}/nodes", ReturnNodes)
  router.HandleFunc("/cex/{source}/nodes/{filter}", ReturnSpecNodes)
	router.HandleFunc("/", TestIndex)
	log.Println("Listening at" + serverIP + "...")
	log.Fatal(http.ListenAndServe(serverIP, router))
}

func TestIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Online!")
}
