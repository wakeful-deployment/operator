package main

import (
  "fmt"
  "os"
  "net/http"
  "encoding/json"
  "syscall"
  "strconv"
  "time"
)

type ConsulKey struct {
  Key string
  ModifyIndex int
}

func handleConsulKeys(resps []ConsulKey, currentIndex int) {
  for _, resp := range resps {
    if (resp.ModifyIndex > currentIndex) {
      fmt.Println("The key is more advanced than we currently know about")
    }
  }
}

func main () {
  hostname, _ := os.LookupEnv("HOSTNAME")
  var index = 0

  for {
    var url = fmt.Sprintf("http://192.168.99.100:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d", hostname, index)
    resp, err := http.Get(url)

    if err != nil {
      fmt.Println("There was an error connecting to consul.")
      syscall.Exit(1)
    }

    switch resp.StatusCode {
    case 200:
      newIndex, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

      if err != nil {
        fmt.Println("There was no 'X-Consul-Index' header...")
        syscall.Exit(1)
      }

      var consulKeys []ConsulKey
      err = json.NewDecoder(resp.Body).Decode(&consulKeys)

      if err != nil {
        fmt.Println("There was an error decoding consule response json that we need to handle.")
        syscall.Exit(1)
      }

      handleConsulKeys(consulKeys, index)

      index = newIndex
    case 404:
      fmt.Println("No keys in keyspace were found. Trying again in one second...")
      time.Sleep(time.Second)
    default:
      fmt.Println("Response was neither 200 nor 404. Something is wrong...")
      time.Sleep(time.Second)
    }
  }
}
