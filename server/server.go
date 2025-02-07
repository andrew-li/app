package main

import (
  "fmt"
  "net/http"
  "flag"    
  "log"
  //"strings"
  //"go/build"
  //"path/filepath"
  //"html/template"   
  "database/sql"
  _ "github.com/go-sql-driver/mysql" 
  "github.com/gorilla/sessions"   
  "net/url"
  "strconv"
  "strings"
)

//"constant" variables to be used throughout the program
const (
  //database configuration information
  DB_USER = "root"
  DB_PASSWORD = ""
  DB_NAME = "virtual_arm"

  //int to represent an invalid selection
  INVALID_INT = -9999
)

//variables to be used throughout the program
var (
  //cookie information
  store = sessions.NewCookieStore([]byte("a-secret-string"))

  //server address information
  addr = flag.String("addr", ":8080", "http service address")  
)

func main() {
  flag.Parse()

  //open the database connection
  var db = initializeDB()
  defer db.Close() //defer closing the connection

  //create the game room
  //var room = createGameRoom(1)
  //go room.run()


  //serve static assets
  http.Handle("/", http.FileServer(http.Dir("../pub/build")))


  //TODO: make urls more RESTful


  //routes in auth.go

  //GET:
  //getUserInfo                                 profile/
  //POST:
  //updateUserInfo                              profile/                       body: {"bio" : bio, "avatar_link" : avatar_link}
  http.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        getUserInfoHandler(w, r, db, store)

        break 
      case "POST":

        updateUserInfoHandler(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

  //POST:
  //createUser                                   users/                         body: {"username" : username, "password" : password, "firstname" : firstname, "lastname" : lastname}
  http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":
        break 
      case "POST":

        createUserHandler(w, r, db)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

  //POST:
  //authenticate                                 authenticate/                  body: {"username" : username, "password" : password}  
  http.HandleFunc("/authenticate/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":
        break 
      case "POST":

        loginHandler(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  }) 


  //routes in forum_threads.go

  //GET:
  //getForumThreadByThreadId                     thread/ id
  //POST:
  //upvoteForumThread                            thread/XidX/?upvote=true       //id in url so don't need body
  //downvoteForumThread                          thread/XidX/?downvote=true     //id in url so don't need body
  //PUT:
  //editForumThread                              thread/                        body: {"title" : title, "body" : body, "link" : link, "tag" : tag}
  //DELETE:
  //deleteForumThread                            thread/ id
  http.HandleFunc("/thread/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          threadId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          getForumThread(w, r, db, 0, INVALID_INT, INVALID_INT, threadId)
        } else {
          //error
        }
        
        break 
      case "POST":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for userid parameter in url
        option := int(0)
        if val, ok := m["upvote"]; ok { //TODO: case insensitve match
          if(val[0] == "true") {
            option = 1
          }
        } else if val, ok := m["downvote"]; ok {
          if(val[0] == "true") {
            option = -1
          }
        }

        if option == 1 || option == -1 {

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 4 {
            threadId, err := strconv.Atoi(s[2])

            if err != nil {
              //error
            }

            scoreForumThread(w, r, db, store, option, threadId)
  
          } else {
            //error
          }

        }

        break  
      case "PUT":

        editForumThread(w, r, db, store)

        break  
      case "DELETE":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          threadId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          deleteForumThread(w, r, db, store, threadId)

        } else {
          //error
        }

        break  
      default:
        break
    }
  })

  //GET:
  //getForumThreadsByLoggedInUserIdByRating      profilethreads/ ? sortby = XXX & pagenumber = XXX
  //getForumThreadsByLoggedInUserIdByTime        profilethreads/ ? sortby = XXX & pagenumber = XXX  
  http.HandleFunc("/profilethreads/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        getForumThreadProtected(w, r, db, store, sortBy, pageNumber)

        break 
      case "POST":
        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

  //GET:  
  //getForumThreadsByUserIdByRating              threads/ ? userid = XXX & sortby = XXX & pagenumber = XXX
  //getForumThreadsByUserIdByTime                threads/ ? userid = XXX & sortby = XXX & pagenumber = XXX
  //getForumThreadsByRating                      threads/ ? sortby = XXX & pagenumber = XXX
  //getForumThreadsByTime                        threads/ ? sortby = XXX & pagenumber = XXX
  //POST:
  //createForumThread                            threads/                       body: {"title" : title, "body" : body, "link" : link, "tag" : tag}
  http.HandleFunc("/threads/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        //select what to query by
        if val, ok := m["userid"]; ok { //by userid
          userId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          getForumThread(w, r, db, 1, sortBy, pageNumber, userId)

        } else { //by all

          getForumThread(w, r, db, 2, sortBy, pageNumber, INVALID_INT)

        }

        break 
      case "POST":

        createForumThread(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

  //GET:
  //trending       trending/
  http.HandleFunc("/trending/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      popularThreads(w, r, db)

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})


  //routes in thread_posts.go

  //GET:
  //getThreadPostByPostId                        post/ id
  //POST:
  //upvoteThreadPost                             post/XidX/?upvote=true         //id in url so don't need body
  //downvoteThreadPost                           post/XidX/?downvote=true       //id in url so don't need body
  //PUT:
  //editThreadPost                               post/                          body: {"thread_id" : threadId, "contents" : contents}
  //DELETE:
  //deleteThreadPost                             post/ id
  http.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          postId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          getThreadPost(w, r, db, 0, INVALID_INT, INVALID_INT, postId)

        } else {
          //error
        }

        break 
      case "POST":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for userid parameter in url
        option := 0
        if val, ok := m["upvote"]; ok {
          if(val[0] == "true") {
            option = 1
          }
        } else if val, ok := m["downvote"]; ok {
          if(val[0] == "true") {
            option = -1
          }
        }

        if option == 1 || option == -1 {

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 4 {
            threadId, err := strconv.Atoi(s[2])

            if err != nil {
              //error
            }

            scoreThreadPost(w, r, db, store, option, threadId)
  
          } else {
            //error
          }

        }

        break  
      case "PUT":

        editThreadPost(w, r, db, store)

        break  
      case "DELETE":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          postId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          deleteThreadPost(w, r, db, store, postId)

        } else {
          //error
        }

        break  
      default:
        break
    }
  })

  //GET:
  //getThreadPostsByThreadIdByRating             posts / ? threadid = XXX & sortby = rating & pagenumber = XXX
  //getThreadPostsByThreadIdByTime               posts / ? threadid = XXX & sortby = creationtime & pagenumber = XXX
  //getThreadPostsByUserIdByRating               posts / ? userid = XXX & sortby = rating & pagenumber = XXX
  //getThreadPostsByUserIdByTime                 posts / ? userid = XXX & sortby = creationtime & pagenumber = XXX
  //getThreadPostsByRating                       posts / ? sortby = rating & pagenumber = XXX
  //getThreadPostsByTime                         posts / ? sortby = creationtime & pagenumber = XXX
  //POST:
  //createThreadPost                             posts/                         body: {"thread_id" : threadId, "contents" : contents}
  http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        //select what to query by
        if val, ok := m["threadid"]; ok { //by threadid
          threadId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          getThreadPost(w, r, db, 1, sortBy, pageNumber, threadId)
        } else if val, ok := m["userid"]; ok { //by userid
          userId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          getThreadPost(w, r, db, 2, sortBy, pageNumber, userId)
        } else { //by all

          getThreadPost(w, r, db, 3, sortBy, pageNumber, INVALID_INT)
        }

        break 
      case "POST":

        createThreadPost(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })


  //routes in users.go

  //GET:
  //getUserInfoByUserId                          user / 1
  //getUserInfoByUsername                        user / ? username = XXX
  http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        if val, ok := m["username"]; ok { //if username can be parsed from url

          username := val[0]

          getUserInfo(w, r, db, 1, INVALID_INT, username)

        } else { //else if username cannot be parsed from url

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 3 { //if path can be divided into 3 parts

            userId, err := strconv.Atoi(s[2]) //convert parsed user id into an int

            if err != nil { //if parsed user id could not be converted into an int
              //error
            }
            getUserInfo(w, r, db, 0, userId, "")
          } else {
            //error
          }
        }

        break 
      case "POST":
        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })


  //routes in friend.go

  http.HandleFunc("/friend/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      GetFriendsList(w, r, db)
      break 
    case "POST":
      m, _ := url.ParseQuery(r.URL.RawQuery)
      if val, ok := m["action"]; ok {
        if val[0] == "add" {
          addFriend(w, r, db)
        } else if val[0] == "remove" {
          removeFriend(w, r, db)
        } else {
          //error
        }
      }
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})

//routes in search.go

//GET:
//search      search/?title=hello&sortby=rating&pagenumber=1
http.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      m, _ := url.ParseQuery(r.URL.RawQuery)

      //look for title parameter in url
      var title string
      if val, ok := m["title"]; ok {
        if len(val[0]) > 2 {
          title = val[0]
        } else {
          //error
          return
        }
      }

      //look for sortby parameter in url
      sortBy := 0
      if val, ok := m["sortby"]; ok {
        if val[0] == "creationtime" {
          sortBy = 1
        }
      }

      //look for pagenumber parameter in url
      pageNumber := 0
      var err error
      if val, ok := m["pagenumber"]; ok {
        pageNumber, err = strconv.Atoi(val[0])
        if err != nil {
          //error
        }
      }

      searchForForumThreads(w, r, db, sortBy, pageNumber, title)

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})


//routes in messages.go

//GET:
//getMessage                               message/ id 
//DELETE:
//deleteMessage                            message/ id
http.HandleFunc("/message/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      s := strings.Split(r.URL.Path, "/")
      if len(s) == 3 {
        messageId, err := strconv.Atoi(s[2])

        if err != nil {
          //error
        }

        recvMessages(w, r, db, store, 0, INVALID_INT, INVALID_INT, messageId)

      } else {
        //error
      }

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})

//GET:
//getMessage                               messages/?q=sender&sortby=desc&pagenumber=1
//getMessage                               messages/?q=recipient&sortby=desc&pagenumber=1
//POST:
//sendMessage                              messages/                         body: { ... }    
http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      m, _ := url.ParseQuery(r.URL.RawQuery)

      //look for query type parameter in url
      option := 0
      if val, ok := m["q"]; ok {
        if(val[0] == "sender") {
          option = 1
        } else {
          option = 2
        }
      }

      //look for sortby parameter in url
      sortBy := 0
      if val, ok := m["sortby"]; ok {
        if val[0] == "asc" {
          sortBy = 1
        }
      }

      //look for pagenumber parameter in url
      pageNumber := 0
      var err error
      if val, ok := m["pagenumber"]; ok {
        pageNumber, err = strconv.Atoi(val[0])
        if err != nil {
          //error
        }
      }

      recvMessages(w, r, db, store, option, sortBy, pageNumber, INVALID_INT)

      break 
    case "POST":

      createMessage(w, r, db, store)

      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})



  // route for friend_list
  // go h.run()  // place this here because we took out friend_list
  // http.HandleFunc("/friendlist/", func(w http.ResponseWriter, r *http.Request ) {
  //  serveWs(w, r, db)
  // })



  var room = createChatRoom(1)
  go room.run()


  //listen for user chat
  http.HandleFunc("/chat/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("trying to initiate chat")
    chat(w, r, store, room)
  })


  //listen on specified port
  fmt.Println("Server starting")
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe:", err)
  }

  // err := http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
  // if err != nil {
  //   log.Fatal("ListenAndServeTLS: ", err)
  // }
}

//function to open connection with database
func initializeDB() *sql.DB {
  db, err := sql.Open("mysql",  DB_USER + ":" + DB_PASSWORD + "@/" + DB_NAME + "?parseTime=true")
  if err != nil {
    panic(err)
  } 

  return db
}

//handle the chat event which checks if the cookie corresponds to a logged in user and adds the user to the chat room
func chat(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, room *ChatRoom) {

  //check for session to see if client is authenticated
  ok, session := confirmSession(store, "Trying to perform action as an invalid user", w, r)
  if ok == false {
    return
  }

  fmt.Println("New user connected to chat")

  //use the id and username attached to the session to create the player
  chatterHandler := ChatterHandler{Id: session.Values["userid"].(int), Username: session.Values["username"].(string), Room: room}

  chatterHandler.createChatter(w, r)
}





