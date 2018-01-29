package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"strconv"
	"bytes"
	"strings"
	"sync"
	"encoding/json"

	"github.com/gosuri/uiprogress"
)

var base string
var jobs chan string
var responses chan string
var hostName string
var inputFilePath string
var hasOutputFile bool
var outputFilePath string
var outputFile io.Writer
var buffer bytes.Buffer
var parallelism int
var bar *uiprogress.Bar

func main() {
	fmt.Println("Running http requester")

	if len(os.Args) < 2 || os.Args[1] == "--help" {
		printHelp(0, "This script will execute the http request on each line of the file provided against the server provided")
	}

	switch len(os.Args) {
	case 3:
		hostName = os.Args[1]
		inputFilePath = os.Args[2]
		hasOutputFile = false
	case 4:
		hostName = os.Args[1]
		inputFilePath = os.Args[2]
		hasOutputFile = true
		outputFilePath = os.Args[3]
	default:
		printHelp(1, "Wrong arguments. You should provide at least two arguemnts for the execution")
	}

	parallelism = 8

	base = fmt.Sprintf("http://%v", hostName)

	// Test the host to see if it's reachable
	_, err := http.Get(base)
	if err != nil {
		fmt.Println(err.Error())
		panic("Failed to make a example GET request to the server")
	}

	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Println(err)
		panic("Cannot open the input file")
	}
	defer inputFile.Close()

	if hasOutputFile {
		outputFile, err = os.Create(outputFilePath)
		if err != nil {
			fmt.Println(err)
			panic("Cannot create the output file")
		}
	} else {
		outputFile = os.Stdout
	}
	
	chanSize := calculateFileSize()

	jobs = make(chan string, chanSize)
	responses = make(chan string, chanSize)

	fileReader := bufio.NewReader(inputFile)
	
	size := 0
	
	for {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		size = size + 1
		jobs <- line
	}
	close(jobs)

	fmt.Printf("Jobs created, sending %v workers\n", parallelism)
	wg1 := new(sync.WaitGroup)
	for i := 0; i< parallelism; i++ {
		wg1.Add(1)
		go worker(makeHttpRequest, jobs, responses, wg1)
	}

	wg2 := new(sync.WaitGroup)

	for i:=0; i < parallelism; i++ {
		wg2.Add(1)
		go writerWorker(responses, wg2)
	}

	fmt.Println("Waiting for jobs to add to file")

	uiprogress.Start()
	bar = uiprogress.AddBar(size)
	bar.AppendCompleted()
	bar.PrependElapsed()
	
	wg1.Wait()
	close(responses)
	wg2.Wait()

	fmt.Fprintf(outputFile, "%v\n",buffer.String())
}

func makeHttpRequest(path string) string {
	req := fmt.Sprintf("%v%v", base, path)
	res, err := http.Get(req)
	var response string
	js := make(map[string]string)
	if err != nil {
		log.Print(err)
		js[strings.TrimSpace(path)] = err.Error()
		byteResponse, err := json.Marshal(js)
			if err != nil {
				log.Print(err)
				js[strings.TrimSpace(path)] = err.Error()
			} else {
				response = string(append(byteResponse, '\n'))
			}
	} else {
		r, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			js[strings.TrimSpace(path)] = err.Error()
			byteResponse, err := json.Marshal(js)
			if err != nil {
				log.Print(err)
				js[strings.TrimSpace(path)] = err.Error()
			} else {
				response = string(append(byteResponse, '\n'))
			}
		} else {
			js[strings.TrimSpace(path)] = string(r)
			byteResponse, err := json.Marshal(js)
			if err != nil {
				log.Print(err)
				js[strings.TrimSpace(path)] = err.Error()
			} else {
				response = string(append(byteResponse, '\n'))
			}
			res.Body.Close()
		}
	}
	return response
}

func printHelp(exitCode int, message string) {
	fmt.Println(message)
	fmt.Println("e.g.: http_requester yourservername.tld:port file_name [output_file]")
	os.Exit(exitCode)
}

func calculateFileSize() int {
	c1 := exec.Command("cat", inputFilePath)
	c2 := exec.Command("wc", "-l")
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r
	var b2 bytes.Buffer
	c2.Stdout = &b2
	c1.Start()
    c2.Start()
    c1.Wait()
    w.Close()
	c2.Wait()
	
	fileSize , err := strconv.Atoi(strings.TrimSpace(b2.String()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return fileSize
}

func worker(execution func(string) string, entry <-chan string, exit chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range entry {
		exit <- execution(job)
	}
}

func writerWorker(inputChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range inputChan {
		buffer.WriteString(job)
		bar.Incr()
	}
}