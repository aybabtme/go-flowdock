package flowdock

import (
	"encoding/json"
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"net/http"
	"time"
)

// MessagesService handles communication with the messages related methods of
// the Flowdock API.
//
// Flowdock API docs: https://www.flowdock.com/api/messages
type MessagesService struct {
	client *Client
}

// MessagesListOptions specifies the optional parameters to the
// MessageService.List method.
type MessagesListOptions struct {
	Event   string   `url:"event,omitempty"`
	Limit   int      `url:"limit,omitempty"`
	SinceID int      `url:"since_id,omitempty"`
	UntilID int      `url:"until_id,omitempty"`
	Tags    []string `url:"tags,comma,omitempty"`
	TagMode string   `url:"tag_mode,omitempty"`
	Search  string   `url:"search,omitempty"`
}

// Stream the messages for the given flow.
//
// Flowdock API docs: https://flowdock.com/api/streaming and
// https://www.flowdock.com/api/messages
func (s *MessagesService) Stream(token, org, flow string) (chan Message, *eventsource.EventSource, error) {
	retryDuration := 3 * time.Second

	u := fmt.Sprintf("flows/%v/%v?access_token=%v", org, flow, token)

	req, err := s.client.NewStreamRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	messageCh := make(chan Message)
	es := eventsource.New(req, retryDuration)

	go func() {
		defer es.Close()
		for {
			event, err := es.Read()

			if err != nil {
				s.client.Log.Printf("failed to read Stream eventsource: %v", err)
				return
			}

			m := new(Message)
			err = json.Unmarshal([]byte(event.Data), m)
			if err != nil {
				s.client.Log.Printf("bad JSON data from Stream eventsource: %v", err)
				return
			}
			messageCh <- *m
		}
	}()

	return messageCh, es, err
}

// List of the messages for the given flow.
//
// Flowdock API docs: https://www.flowdock.com/api/messages
func (s *MessagesService) List(org, flow string, opt *MessagesListOptions) ([]Message, *http.Response, error) {
	u := fmt.Sprintf("flows/%v/%v/messages", org, flow)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var messages []Message
	resp, err := s.client.Do(req, &messages)
	if err != nil {
		return nil, resp, err
	}

	return messages, resp, err
}

// Get a single message by ID.
//
// Flowdock API docs: https://www.flowdock.com/api/messages
func (s *MessagesService) Get(org, flowName string, id int) (*Message, *http.Response, error) {
	u := fmt.Sprintf("flows/%v/%v/messages/%d", org, flowName, id)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	message := new(Message)
	resp, err := s.client.Do(req, message)
	if err != nil {
		return nil, resp, err
	}

	return message, resp, err
}

type MessagesEditOptions struct {
	Content string   `url:"content,omitempty"`
	Tags    []string `url:"tags,comma,omitempty"`
}

func (s *MessagesService) Edit(org, flowName string, id int, opt *MessagesEditOptions) (*http.Response, error) {
	u := fmt.Sprintf("/flows/%s/%s/messages/%d", org, flowName, id)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest("PUT", u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *MessagesService) Delete(org, flowName string, id int) (*http.Response, error) {
	u := fmt.Sprintf("/flows/%s/%s/messages/%d", org, flowName, id)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(req, nil)
}

// MessagesCreateOptions specifies the optional parameters to the
// MessageService.Create method.
type MessagesCreateOptions struct {
	FlowID           string   `url:"flow,omitempty"`
	MessageID        int      `url:"message,omitempty"`
	Event            string   `url:"event,omitempty"`
	Content          string   `url:"content,omitempty"`
	Tags             []string `url:"tags,comma,omitempty"`
	UUID             string   `url:"uuid,omitempty"`
	ExternalUserName string   `url:"external_user_name,omitempty"`
	Subject          string   `url:"subject,omitempty"`
	FromAddress      string   `url:"from_address,omitempty"`
	Source           string   `url:"source,omitempty"`
}

// CreateComment for the specified organization
//
// Flowdock API docs: https://www.flowdock.com/api/messages
func (s *MessagesService) CreateComment(opt *MessagesCreateOptions) (*Message, *http.Response, error) {
	u := "comments"

	u, err := addOptions(u, opt)
	req, err := s.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	message := new(Message)
	resp, err := s.client.Do(req, message)
	if err != nil {
		return nil, resp, err
	}

	return message, resp, err
}

// Create a message for the specified organization
//
// Flowdock API docs: https://www.flowdock.com/api/messages
func (s *MessagesService) Create(opt *MessagesCreateOptions) (*Message, *http.Response, error) {
	u := "messages"

	u, err := addOptions(u, opt)
	req, err := s.client.NewRequest("POST", u, nil)
	if err != nil {
		return nil, nil, err
	}

	message := new(Message)
	resp, err := s.client.Do(req, message)
	if err != nil {
		return nil, resp, err
	}

	return message, resp, err
}

// Message represents a Flowdock chat message.
type Message struct {
	ID               *int             `json:"id,omitempty"`
	FlowID           *string          `json:"flow,omitempty"`
	Sent             *Time            `json:"sent,omitempty"`
	UserID           *string          `json:"user,omitempty"`
	Event            *string          `json:"event,omitempty"`
	RawContent       *json.RawMessage `json:"content,omitempty"`
	MessageID        *int             `json:"message,omitempty"`
	Tags             *[]string        `json:"tags,omitempty"`
	UUID             *string          `json:"uuid,omitempty"`
	ExternalUserName *string          `json:"external_user_name,omitempty"`
	App              *string          `json:"app,omitempty"` // deprecated
}

// Content of a Message
//
// It can be a MessageContent, CommentContent, etc. Depends on the Event
func (m *Message) Content() (content Content) {

	switch *m.Event {
	case "message":
		content = new(MessageContent)
	case "comment":
		content = &CommentContent{}
	case "vcs":
		content = &VcsContent{}
	default:
		content = new(JsonContent)
	}

	if err := json.Unmarshal([]byte(*m.RawContent), &content); err != nil {
		panic(err.Error())
	}

	return content
}
