package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "micrified/sql.auth"
  "micrified/sql.driver"
  "net/http"
  "os"
  //"net/url"
)

// Usage describes how to use the program
const Usage         = "<server-config.json>"
const queryID       = "id"

// Global driver handle
var D driver.Driver

// Global server configuration
var C Config

// Global session manager
var S auth.SessionManager

// Defines: Server configuration
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

type Table struct {
  recordTable string
  contentTable string
}

func (t *Table) RecordTable() string {
  return t.recordTable
}

func (t *Table) ContentTable() string {
  return t.contentTable
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

// Authentication 

// onPostLogin handles a POST request for authorization 
func onPostLogin(w http.ResponseWriter, r *http.Request, z driver.Tables) {
  status, buffer := http.StatusOK, bytes.Buffer{}

  // Receive login. Verify hash. Return SID
  body, err := driver.User

  w.Header.Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("LOGIN: %d bytes\n", buffer.Len())
}

// onGetStatic handles a GET request for static pages
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
func onGetStatic(w http.ResponseWriter, r *http.Request, z driver.Tables) {
  status, buffer := http.StatusOK, bytes.Buffer{}
  
  body, err := driver.StaticPage(&D, r.URL.Query().Get(queryID), z)
  if nil != err {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    buffer.Write(body)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("GET (static): %d bytes\n", buffer.Len())
}

// onGet handles a GET request for a single SQLType T
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
// p: Template type configured with criteria
func onGet [T driver.SQLType[T], P interface{*T;driver.Queryable}] (w http.ResponseWriter, r *http.Request, z driver.Tables, p P) {
  status, buffer := http.StatusOK, bytes.Buffer{}
  item, err := driver.Row[T,P](&D, p, z)
  if nil != err {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&item)
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("GET: %d bytes\n", buffer.Len())
}

// onGetList handles a GET request for all SQLType T
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
// p: Template type (unused)
func onGetList [T driver.SQLType[T], P interface{*T;driver.Queryable}] (w http.ResponseWriter, r *http.Request, z driver.Tables, p P) {
  status, buffer := http.StatusOK, bytes.Buffer{}
  list, err := driver.Rows[T,P](&D, p, z)
  if nil != err {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&list)
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("GET (List): %d bytes\n", buffer.Len())
}

// onPost handles a POST request for an SQLType T
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
// p: Template type containing SQL data to be entered
func onPost [T driver.SQLType[T], P interface{*T;driver.Queryable}] (w http.ResponseWriter, r *http.Request, z driver.Tables, p P) {
  status, buffer := http.StatusOK, bytes.Buffer{}
  var item T

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, p)
  }
  if nil == err {
    item, err = driver.Insert[T,P](&D, p, z)
  } else {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  if err != nil {
    status = http.StatusBadRequest
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&item)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("POST: %d bytes\n", buffer.Len())
}

// onPut handles a PUT request for an SQLType T
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
// p: Template type containing SQL data to be updated
func onPut [T driver.SQLType[T], P interface{*T;driver.Queryable}] (w http.ResponseWriter, r *http.Request, z driver.Tables, p P) {
  status, buffer := http.StatusOK, bytes.Buffer{}
  var item T

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, p)
  }
  if nil == err {
    item, err = driver.Update[T,P](&D, p, z)
  } else {
    status = http.StatusNotFound
    buffer.WriteString("Failed resource request: " + err.Error())
  }

  if nil != err {
    status = http.StatusNotFound
    buffer.WriteString("Failed resource request: " + err.Error())
  } else {
    json.NewEncoder(&buffer).Encode(&item)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(buffer.Bytes())
  log.Printf("PUT: %d bytes\n", buffer.Len())
}

// onDelete handles a DELETE request for an SQLType T
// w: ResponseWriter structure
// r: Pointer to request structure
// z: Database tables structure
// p: Template type containing SQL data to be deleted
func onDelete [T driver.SQLType[T], P interface{*T;driver.Queryable}] (w http.ResponseWriter, r *http.Request, z driver.Tables, p P) {
  status, buffer := http.StatusOK, bytes.Buffer{}

  // Unmarshal the JSON encoded content
  body, err := ioutil.ReadAll(r.Body)
  if nil == err {
    err = json.Unmarshal(body, p)
  }
  if nil == err {
    err = driver.Delete[T,P](&D,p,z)
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
  log.Printf("DELETE: %d bytes\n", buffer.Len())
}

// handleLogin handles HTTP requests to the login interface
func handleLogin (w http.ResponseWriter, r *http.Request) {
  z := Table{recordTable: "users", contentTable: "credentials"}
  if http.MethodPost == r.Method {
    onPostLogin(w, r, &z)
  }
}

// handleStatic handles HTTP requests to any kind of static page
func handleStatic (w http.ResponseWriter, r *http.Request) {
  z := Table{recordTable: "static_pages", contentTable: "page_content"}
  if http.MethodGet == r.Method {
    onGetStatic(w, r, &z)
  }
  // TODO: Handle MethodPost, MethodPut, etc.
}

// handleBlogs: HTTP handler for requests to the /blogs subdomain
func handleBlogs(w http.ResponseWriter, r *http.Request) {
  z := &Table{recordTable: "blog_pages", contentTable: "page_content"}
  page, query := driver.Page{}, r.URL.Query()
  switch r.Method {
  case http.MethodGet:
    page.ID = query.Get(queryID)
    if len(page.ID) > 0 {
      onGet[driver.Page](w, r, z, &page)
    } else {
      onGetList[driver.Page](w, r, z, &page)
    }
  case http.MethodPost:
    onPost[driver.Page](w, r, z, &page)
  case http.MethodPut:
    onPut[driver.Page](w, r, z, &page)
  case http.MethodDelete:
    onDelete[driver.Page](w, r, z, &page)
  }
}

// handlePastes: HTTP handler for requests to the /pastes subdomain
func handlePastes (w http.ResponseWriter, r *http.Request) {
  z := &Table{recordTable: "paste_pages", contentTable: "page_content"}
  paste, query := driver.Paste{}, r.URL.Query()
  switch r.Method {
  case http.MethodGet:
    paste.ID = query.Get(queryID)
    if len(paste.ID) > 0 {
      onGet[driver.Paste](w, r, z, &paste)
    } else {
      onGetList[driver.Paste](w, r, z, &paste)
    }
  case http.MethodPost:
    onPost[driver.Paste](w, r, z, &paste)
  case http.MethodPut:
    onPut[driver.Paste](w, r, z, &paste)
  case http.MethodDelete:
    onDelete[driver.Paste](w, r, z, &paste)
  }
}

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

  // Initialize session handling
  S = auth.CreateSessionManager[string](1 * time.Hour)

  // Register handler: /login
  http.HandleFunc("/login", handleLogin)

  // Register handler: /pages
  http.HandleFunc("/static", handleStatic)

  // Register handler: /blogs
  http.HandleFunc("/blogs", handleBlogs)

  // Register handler: /pastes
  http.HandleFunc("/pastes", handlePastes)

  // Listen and serve
  addr := fmt.Sprintf("%s:%s", C.Host, C.Port)
  fmt.Printf("Listening at: %s\n", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}
