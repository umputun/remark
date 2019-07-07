package notify

import (
	"context"
	"fmt"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

// EmailParams contain settings for email set up
type EmailParams struct {
	Server    string
	Port      int
	From      string
	Username  string
	Password  string
	KeepAlive time.Duration
}

// Email implements notify.Destination for email
type Email struct {
	server    string
	port      int
	from      string
	username  string
	password  string
	keepAlive time.Duration
	sendChan  chan *gomail.Message
	errChan   chan error
}

const emailConnectionKeepAlive = 30 * time.Second

//NewEmail makes email object for notifications and run sending daemon
func NewEmail(params EmailParams) (*Email, error) {

	res := Email{
		server:    params.Server,
		port:      params.Port,
		from:      params.From,
		username:  params.Username,
		password:  params.Password,
		keepAlive: params.KeepAlive,
		sendChan:  make(chan *gomail.Message),
		errChan:   make(chan error),
	}

	if res.keepAlive == 0 {
		res.keepAlive = emailConnectionKeepAlive
	}

	log.Printf("[DEBUG] create new email notifier for server %s with user %s, keepalive=%s",
		res.server, res.username, res.keepAlive)

	// Test connection before starting a daemon.
	tmpConn, err := gomail.NewDialer(res.server, res.port, res.username, res.password).Dial()
	if err != nil {
		return &res, errors.Wrapf(err, "error establishing test connecting to '%s':%d with username '%s'",
			res.server, res.port, res.username)
	}
	if err = tmpConn.Close(); err != nil {
		return &res, errors.Wrapf(err, "error closing test connection to %s:%d",
			res.server, res.port)
	}

	// Activate server goroutine.
	go res.activate()
	// TODO: server goroutine never dies! This exiter should be run once first time we get context
	// Close the channel to stop the mail daemon
	// go func() { <-ctx.Done(); close(res.sendChan) }()

	return &res, nil
}

// Send email from request to address in settings
func (e *Email) Send(ctx context.Context, req request) error {
	log.Printf("[DEBUG] send notification via %s, comment id %s", e, req.comment.ID)

	messageBody := prepareBody(req)

	// Create message.
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	// TODO: figure out where "to" addresses come from
	//m.SetAddressHeader("To", req.Address, req.Name)
	m.SetHeader("Subject", fmt.Sprintf("New comment for \"%s\"", req.comment.PostTitle))
	m.SetBody("text/html", messageBody)

	// Wait for ability to send message and return error from error channel after sending it.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.sendChan <- m:
		return <-e.errChan
	}
}

func (e *Email) String() string {
	return fmt.Sprintf("email: '%s'@'%s':%d", e.username, e.server, e.port)
}

//prepareBody generates email message text based on request
func prepareBody(req request) string {
	from := req.comment.User.Name
	if req.comment.ParentID != "" {
		from += " → " + req.parent.User.Name
	}
	link := fmt.Sprintf("↦ [original comment](%s)", req.comment.Locator.URL+uiNav+req.comment.ID)
	if req.comment.PostTitle != "" {
		link = fmt.Sprintf("↦ [%s](%s)", req.comment.PostTitle, req.comment.Locator.URL+uiNav+req.comment.ID)
	}
	body := fmt.Sprintf("%s\n\n%s\n\n%s", from, req.comment.Orig, link)
	return body
}

func (e *Email) activate() {
	d := gomail.NewDialer(e.server, e.port, e.username, e.password)

	var s gomail.SendCloser
	var err error
	open := false
	for {
		select {
		case m, ok := <-e.sendChan:
			if !ok {
				close(e.errChan)
				return
			}
			if !open {
				if s, err = d.Dial(); err != nil {
					// Problems with connection, returning error and considering connection not established.
					e.errChan <- errors.Wrapf(err, "error connecting to %s:%d with username %s",
						e.server, e.port, e.username)
					break
				}
				open = true
			}
			err = gomail.Send(s, m)
			e.errChan <- errors.Wrapf(err, "error sending to %s:%d", e.server, e.port)
		// Close the connection to the SMTP server if no email was sent in the keepAlive period.
		case <-time.After(e.keepAlive):
			if open {
				if err := s.Close(); err != nil {
					// Problems with closing connection, considering connection still established.
					log.Printf("[WARN] error closing connection with %s:%d with username %s: %s",
						e.server, e.port, e.username, err)
					break
				}
				open = false
			}
		}
	}
}
