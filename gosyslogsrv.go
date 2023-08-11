/*
2023-03-22 log server
2023-07-11 log server
2023-07-14 log server
yumin wu

//SQL:uuid INTEGER PRIMARY KEY AUTOINCREMENT,time timestamp NULL DEFAULT CURRENT_TIMESTAMP,host,data

go get github.com/mattn/go-sqlite3
*/

package main

import (
  "flag"
  "fmt"
  "time"
  "net"
  "os"
  "os/signal"
  "strconv"
  "syscall"
  "runtime"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

const AppVersion = "0.0.1 beta uwmy"

var DEBUG = false
var PROGRESS = true

var DB *sql.DB
var Stmt *sql.Stmt
var DR0 string     //setup work directory
var DB0 string     //setup dbname
var TB0 string     //setup table name
var PORT0 int      //setup syslog port
var DbMode string  //sqlite3 Journal mode DELETE,TRUNCATE,PERSIST,MEMORY,WAL,OFF

var SIGTERM int = 0

var listener net.PacketConn
var err error

func init() {
  d, _ := os.Getwd();
  flag.StringVar(&DR0, "d", d, "work directory")
  flag.StringVar(&DB0, "f", "db_" + time.Now().Format("2006-01-02_150405") + ".sl3", "sqlite3 file")
  flag.StringVar(&TB0, "t", "log", "database table name")
  flag.IntVar(&PORT0, "p", 514, "syslog server port")
  flag.StringVar(&DbMode, "dbmode", "WAL", "sqlite3 Journal mode DELETE,TRUNCATE,PERSIST,MEMORY,WAL,OFF")
  flag.BoolVar(&DEBUG, "debug", false, "debug mode")
  flag.BoolVar(&PROGRESS, "progress", false, "show progress")
}

func main() {
  SetupCtrlC()
  version := flag.Bool("v", false, "prints current version")
  flag.Parse()
  if DEBUG { PROGRESS = false }

  if *version {
    fmt.Println(AppVersion)
    os.Exit(0)
  }

fmt.Println("DR0=", DR0)
fmt.Println("DB0=", DB0)
fmt.Println("TB0=", TB0)
fmt.Println("PORT0=", PORT0)
fmt.Println("Debug=", DEBUG)
fmt.Println("DbMode=", DbMode)
fmt.Println("PROGRESS=", PROGRESS)
 
  err := ConnectDatabase()
  checkErr(err)

  listener, err = net.ListenPacket("udp", "0.0.0.0:" + strconv.Itoa(PORT0))
  if err != nil { fmt.Println(err.Error()) }

  defer listener.Close()
  fmt.Printf("Log server start and listening on %d.\n", PORT0)

  syscall.Chdir(DR0)

  d, _ := syscall.Getwd()
  fmt.Printf("Change work directory: %s \n", d)

  sql := fmt.Sprintf("CREATE TABLE if not exists %s(i INTEGER PRIMARY KEY AUTOINCREMENT,t timestamp NULL DEFAULT CURRENT_TIMESTAMP,h,d)", TB0)
  _, err = DB.Exec(sql)
  _, err = DB.Exec(fmt.Sprintf("PRAGMA journal_mode=%s", DbMode))
  checkErr(err)
  InsertLogStmt()
  fmt.Printf("Open Sqlite3: %s@%s\n", TB0, DB0)

  for {
    for SIGTERM > 0 { time.Sleep(time.Second) }

    buf := make([]byte, 4096)
    n, addr, err := listener.ReadFrom(buf)
    if err != nil {  continue }
    go serve(listener, addr, buf[:n])
  }
  DB.Close()
}

func serve(listener net.PacketConn, addr net.Addr, buf []byte) {

  i := 0
  for i < 100 {     
    i = i + 1 
    _, err := Stmt.Exec(addr.String(), string(buf))

    if err == nil { break }
    time.Sleep(200 * time.Millisecond)
  }
  if PROGRESS { fmt.Print(" ", runtime.NumGoroutine(), ":",i) }
  if DEBUG { fmt.Print(runtime.NumGoroutine(), i, addr, string(buf)) }
}

func ConnectDatabase() error {
  db, err := sql.Open("sqlite3", DB0)
  if err != nil { return err }
  DB = db
  return nil
}

func InsertLogStmt() error {
  stmt, err := DB.Prepare(fmt.Sprintf("INSERT INTO %s(h, d) values(?,?)", TB0))
  checkErr(err)
  Stmt = stmt
  return nil
}

func checkErr(err error) {
  if err != nil { panic(err) }
}

func Exists(path string) bool {
  _, err := os.Stat(path)
  if err != nil {
    if os.IsExist(err){ return true }
    return false
  }
  return true
}

func IsDirectory(path string) bool {
  fileInfo, err := os.Stat(path)
  if err != nil {  return false }
  return fileInfo.IsDir()
}

func SetupCtrlC() {
  c := make(chan os.Signal, 2)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)

  go func() {
    <-c
    SIGTERM = 3

    listener.Close()

    fmt.Print("\nStop Process: ", runtime.NumGoroutine(), "\n")
    for runtime.NumGoroutine() >4 {
      fmt.Print(".")
      time.Sleep(time.Second)
    }
    fmt.Println("End Process: ", runtime.NumGoroutine())
    os.Exit(0)
  }()
}

