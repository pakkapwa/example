package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

const (
	FACEBOOK_API = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"
	IMAGE        = "http://37.media.tumblr.com/e705e901302b5925ffb2bcf3cacb5bcd/tumblr_n6vxziSQD11slv6upo3_500.gif"
)

type Callback struct {
	Object string `json:"object,omitempty"`
	Entry  []struct {
		ID        string      `json:"id,omitempty"`
		Time      int         `json:"time,omitempty"`
		Messaging []Messaging `json:"messaging,omitempty"`
	} `json:"entry,omitempty"`
}

type Messaging struct {
	Sender    User    `json:"sender,omitempty"`
	Recipient User    `json:"recipient,omitempty"`
	Timestamp int     `json:"timestamp,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type User struct {
	ID string `json:"id,omitempty"`
}

type Message struct {
	MID        string `json:"mid,omitempty"`
	Text       string `json:"text,omitempty"`
	QuickReply *struct {
		Payload string `json:"payload,omitempty"`
	} `json:"quick_reply,omitempty"`
	Attachments *[]Attachment `json:"attachments,omitempty"`
	Attachment  *Attachment   `json:"attachment,omitempty"`
}

type Attachment struct {
	Type    string  `json:"type,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

type Response struct {
	Recipient User    `json:"recipient,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type Payload struct {
	URL string `json:"url,omitempty"`
}

func VerificationEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------------------------------------------VerificationEndpoint-Test 5555-------------------------------------------------------------------") 
	challenge := r.URL.Query().Get("hub.challenge")
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	ver_token := os.Getenv("VERIFY_TOKEN")
	fmt.Println("--------------------------------challenge-----------"+challenge+"-----------------------------------------------------")
	fmt.Println("--------------------------------ver_token-------------"+ver_token+"-----------------------------------------------------")
	if mode != "" && token == ver_token {
		w.WriteHeader(200)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Error, wrong validation token"))
	}
}

func ProcessMessage(event Messaging) {
	fmt.Println("-----------------------------------------------------ProcessMessage-------------------------------------------------------------------")
	fmt.Println("-----------------------------------------------------"+event.Sender.ID+"-------------------------------------------------------------------")
	client := &http.Client{}
	response := Response{
		Recipient: User{
			ID: event.Sender.ID,
		},
		Message: Message{
			Text:"champ ja",
		},
	}


	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(&response)
	url := fmt.Sprintf(FACEBOOK_API, os.Getenv("PAGE_ACCESS_TOKEN"))
	req, err := http.NewRequest("POST", url, body)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func MessagesEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------------------------------------------MessagesEndpoint 55-------------------------------------------------------------------")

	// j, _ := json.MarshalIndent(r.Body, "", " ")
	// fmt.Println(string(j))

	var callback Callback
	err := json.NewDecoder(r.Body).Decode(&callback)
	if err != nil {
		log.Println("err decode : ", err)
	}


	j, _ := json.Marshal(callback)
	fmt.Println(string(j))
	
	if callback.Object == "page" {
		ctx := appengine.NewContext(r)
		cli := urlfetch.Client(ctx)

		url := fmt.Sprintf(FACEBOOK_API, os.Getenv("PAGE_ACCESS_TOKEN"))

		request_message := Response{
			Recipient: User{
				ID: callback.Entry[0].Messaging[0].Sender.ID,
			},
			Message: Message{
				Text:"champ ja",
			},
		}

		jsonRequestPostBackMessage, _ := json.MarshalIndent(request_message, "", " ")
		byteRequestPostBackMessage := []byte(jsonRequestPostBackMessage)
		requestReaderPostBackMessage := bytes.NewReader(byteRequestPostBackMessage)

		cli.Post(url, "application/json", requestReaderPostBackMessage)
		

		// ProcessMessage(callback.Entry[0].Messaging[0])
		// for _, entry := range callback.Entry {
		// 	for _, event := range entry.Messaging {
		// 		ProcessMessage(event)
		// 	}
		// }
	} 
	w.WriteHeader(200)

}

func main() {
	port:=os.Getenv("PORT")
	r := mux.NewRouter()
	r.HandleFunc("/webhook", VerificationEndpoint).Methods("GET")
	r.HandleFunc("/webhook", MessagesEndpoint).Methods("POST")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
