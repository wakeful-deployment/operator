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

// Find keys that are present in first slice that are not present in second
func keyDiff(keySlice1 []ConsulKey, keySlice2 []ConsulKey) []ConsulKey {
  var diff []ConsulKey

  for _, keyInFirst := range keySlice1 {
    var present = true
    for _, keyInSecond := range keySlice2 {
      if keyInFirst.Key == keyInSecond.Key {
        present = false
        break
      }
    }

    if present {
      diff = append(diff, keyInFirst)
    }
  }

  return diff
}

func handleStateChange(previousState KeyState, newState KeyState) {
  // TODO: This find all keys in namespace that differ.
  // We want to only find 'app' keys
  removedKeys := keyDiff(previousState.Keys, newState.Keys)
  addedKeys := keyDiff(newState.Keys, previousState.Keys)

  fmt.Println("Removed:")
  fmt.Println(removedKeys)
  fmt.Println("Added:")
  fmt.Println(addedKeys)

  out, err := exec.Command("sh","-c",cmd).Output()
}

func main () {
  hostname, _ := os.LookupEnv("HOSTNAME")
  state := KeyState{Keys: []ConsulKey{}, Index: 0}

  for {
    newState, err := getNewKeyState(hostname, state)

    if err != nil {
      fmt.Println("There was an error getting the new state")
      syscall.Exit(1)
    }

    handleStateChange(state, newState)

    time.Sleep(time.Second)

    state = newState
  }
}
