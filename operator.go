package main

import (
  "fmt"
  "os"
  "net/http"
  "encoding/json"
  "syscall"
  "strconv"
  "strings"
  "time"
  "errors"
)

type KeyState struct {
  Index int
  Keys []ConsulKey
}

type ConsulKey struct {
  Key string
  ModifyIndex int
}

func handleConsulKeys(keys []ConsulKey, currentIndex int) {
  for _, key := range keys {
    if (strings.HasPrefix(key.Key, "apps")) {
      fmt.Println("HEY")
    }
  }
}

func url(hostname string, index int) string {
  return fmt.Sprintf("http://192.168.99.100:8500/v1/kv/_wakeful/nodes/%s?recurse=true&index=%d", hostname, index)
}

func getNewKeyState(hostname string, state KeyState) (KeyState, error) {
  resp, err := http.Get(url(hostname, state.Index))

  if err != nil {
    var emptyState KeyState
    return emptyState, err
  }

  switch resp.StatusCode {
  case 200:
    index, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

    if err != nil {
      var emptyState KeyState
      return emptyState, err
    }

    var keys []ConsulKey
    err = json.NewDecoder(resp.Body).Decode(&keys)

    if err !=  nil {
      var emptyState KeyState
      return emptyState, err
    }

    state.Keys = keys
    state.Index = index

    return state, nil
  case 404:
    index, err := strconv.Atoi(resp.Header["X-Consul-Index"][0])

    if err != nil {
      var emptyState KeyState
      return emptyState, err
    }
    state.Index = index
    state.Keys = []ConsulKey{}
    return state, nil
  default:
    var emptyState KeyState
    return emptyState, errors.New("non 200/404 response code")
  }
}

func main () {
  hostname, _ := os.LookupEnv("HOSTNAME")
  state := KeyState{Keys: []ConsulKey{}, Index: 0}

  for {
    newState, err := getNewKeyState(hostname, state)

    // handleStateChange(state, newState)

    for _, key := range newState.Keys {
      fmt.Println(key)
    }

    if err != nil {
      fmt.Println("There was an error getting the new state")
      syscall.Exit(1)
    }

    time.Sleep(time.Second)

    state = newState
  }
}
