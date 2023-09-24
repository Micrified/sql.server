package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "micrified/sql.driver"
  "net/http"
  "net/url"
  "os"
)

// Usage describes how to use the program
const Usage         = "<server-config.json>"
const queryID       = "id"

// Global driver handle
var D driver.Driver

// Global server Configuration
var C Config

// Server configuration
type Config struct {
  Database struct {
    UnixSocket string
    Username   string
    Password   string
    Database   string
  }
  Host         string
  Port         string
}

// Configuration parses the given file into a Config structure
func configuration (filepath string) (Config, error) {
  var c Config
  f, err := os.Open(filepath)
  if nil != err {
    return c, fmt.Errorf("Bad configuration %s: %w", filepath, err)
  } else {
    defer f.Close()
  }
  parser := json.NewDecoder(f)
  err = parser.Decode(&c)
  return c, err
}

// onGetStatic handles a GET request for static pages. It draws from
// a hash table (hTable) for page lookup, and a content table (cTable)
// for page data
func onGetStatic(w http.ResponseWriter, hTable, cTable string, query url.Values) {
  status, buffer, id := http.StatusOK, bytes.Buffer{}, query.Get(queryID)
  body, err := D.StaticPage(cTable, hTable, id)
  
  if nil != err {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    buffer.Write(body)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("Static page request (id=\"%s\"): %d bytes\n",
    id, buffer.Len())
}

// onGetPage handles a GET request for paged content. It draws from 
// a record table (rTable) for page lists, and a content table (cTable)
// for page data
func onGetPage (w http.ResponseWriter, rTable, cTable string, query url.Values) {
  status, buffer, id := http.StatusOK, bytes.Buffer{}, query.Get(queryID)
  
  // Case: Request for paged content list
  if len(id) == 0 {
    pages, err := D.IndexedPages(rTable, cTable)
    if nil != err {
      status = http.StatusBadRequest
      buffer.WriteString("Failed resource request: " + err.Error())
    } else {
      json.NewEncoder(&buffer).Encode(&pages)
    }

  // Case: Request for specific page
  } else {
    // TODO: Handle invalid query ID
    page, err := D.IndexedPage(rTable, cTable, id)
    if nil != err {
      status = http.StatusBadRequest
      buffer.WriteString("Failed resource request: " + err.Error())
    } else {
      json.NewEncoder(&buffer).Encode(&page)
    }
  }
  
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("Indexed page request (id=\"%s\"): %d bytes\n", 
    id, buffer.Len())
}

func onPostPage (w http.ResponseWriter, rTable, cTable string, r *http.Request) {
  status, buffer, form := http.StatusOK, bytes.Buffer{}, driver.Page{}
  var page = driver.Page{}

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, &form)
  }
  if nil == err {
    page, err = D.InsertIndexedPage(rTable, cTable, form)
  } else {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  if err != nil {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&page)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("Page post request (): %d bytes\n", buffer.Len())
}

func onPutPage(w http.ResponseWriter, rTable, cTable string, r *http.Request) {
  status, buffer, form := http.StatusOK, bytes.Buffer{}, driver.Page{}
  var page = driver.Page{}

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, &form)
  }
  if nil == err {
    page, err = D.UpdateIndexedPage(rTable, cTable, form)
  } else {
    status = http.StatusNotFound
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  if nil != err {
    status = http.StatusNotFound
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&page)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("Page put request(): %d bytes\n", buffer.Len())
}

func onDeletePage(w http.ResponseWriter, rTable, cTable string, r *http.Request) {
  status, buffer, form := http.StatusOK, bytes.Buffer{}, driver.Page{}

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, &form)
  }
  if nil == err {
    err = D.DeleteIndexedPage(rTable, cTable, form)
  } else {
    status = http.StatusNotFound
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  if nil != err {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  w.Header().Set("Content-Type", "text/plain")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("Page delete request(): %d bytes\n", buffer.Len())
}

// handleStatic handles HTTP requests to any kind of static page
func handleStatic (w http.ResponseWriter, r *http.Request) {
  if http.MethodGet == r.Method {
    onGetStatic(w, "static_pages", "page_content", r.URL.Query())
  }
  // TODO: Handle MethodPost, MethodPut, etc.
}

// handleBlogs: HTTP handler for requests to the /blogs subdomain
func handleBlogs(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case http.MethodGet:
    onGetPage(w, "blog_pages", "page_content", r.URL.Query())
  case http.MethodPost:
    onPostPage(w, "blog_pages", "page_content", r)
  case http.MethodPut:
    onPutPage(w, "blog_pages", "page_content", r)
  case http.MethodDelete:
    onDeletePage(w, "blog_pages", "page_content", r)
  }
}

// handlePastes: HTTP handler for requests to the /pastes subdomain
// func handlePastes (w http.ResponseWriter, r *http.Request) {
// }

func main() {
  var err error
  
  // Argument processing
  if len(os.Args) == 2 {
    C, err = configuration(os.Args[1])
    if nil != err {
      log.Fatal(err)
    }
  } else {
    log.Println("Warning: Using default configuration!")
  }

  // Initialize database driver
  dsn, err := D.Init(C.Database.UnixSocket, C.Database.Username, 
    C.Database.Password, C.Database.Database)
  if nil != err {
    log.Fatal(err.Error())
  } else {
    log.Printf("Connected (DSN = \"%s\")\n", dsn)
    defer D.Stop()
  }

  // Register handler: /pages
  http.HandleFunc("/static", handleStatic)

  // Register handler: /blogs
  http.HandleFunc("/blogs", handleBlogs)

  // Register handler: /pastes
  //http.HandleFunc("/pastes", handlePastes)

  // Listen and serve
  addr := fmt.Sprintf("%s:%s", C.Host, C.Port)
  log.Fatal(http.ListenAndServe(addr, nil))
}
