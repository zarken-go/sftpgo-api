package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/plugins/timeout"
)

var (
	ErrUnauthorized = errors.New(http.StatusText(http.StatusUnauthorized))
	ErrForbidden    = errors.New(http.StatusText(http.StatusForbidden))
)

type apiError struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func (err apiError) Error() string {
	if err.Err != `` {
		return err.Err
	}
	return err.Message
}

type Client interface {
	GetAllUsers() (Users, error)
	GetUsers(input GetUsersInput) (Users, error)
	GetUserQuotaScans() (UserQuotaScans, error)
	StartUserQuotaScan(User User) error
	GetActiveConnections() ([]ConnectionStatus, error)
}

type client struct {
	cli *gentleman.Client
}

func NewClient(URL, Username, Password string) Client {
	cli := gentleman.New()
	cli.URL(URL)
	cli.Use(NewTokenPlugin(URL, Username, Password))
	cli.Use(timeout.All(timeout.Timeouts{
		Request:   15 * time.Second,
		TLS:       5 * time.Second,
		Dial:      5 * time.Second,
		KeepAlive: 15 * time.Second,
	}))

	return &client{cli: cli}
}

type GetUsersInput struct {
	Offset int
	Limit  int
	Order  string
}

func (c *client) GetAllUsers() (Users, error) {
	Input := GetUsersInput{
		Offset: 0,
		Limit:  500,
	}

	var all Users
	for {
		page, err := c.GetUsers(Input)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)

		if len(page) < Input.Limit {
			break
		}
		Input.Offset += Input.Limit
	}
	return all, nil
}

func (c *client) GetUsers(input GetUsersInput) (Users, error) {
	req := c.cli.Request()
	req.Path(`/api/v2/users`)
	if input.Offset > 0 {
		req.SetQuery(`offset`, strconv.Itoa(input.Offset))
	}
	if input.Limit > 0 {
		req.SetQuery(`limit`, strconv.Itoa(input.Limit))
	}
	if input.Order != `` {
		req.SetQuery(`order`, input.Order)
	}

	var users Users
	if err := doJSON(req, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *client) GetUserQuotaScans() (UserQuotaScans, error) {
	req := c.cli.Request()
	req.Path(`/api/v2/quota-scans`)
	var scans UserQuotaScans
	if err := doJSON(req, &scans); err != nil {
		return nil, err
	}
	return scans, nil
}

func (c *client) GetActiveConnections() ([]ConnectionStatus, error) {
	req := c.cli.Request()
	req.Path(`/api/v2/connections`)
	var connections []ConnectionStatus
	if err := doJSON(req, &connections); err != nil {
		return nil, err
	}
	return connections, nil
}

func (c *client) StartUserQuotaScan(User User) error {
	req := c.cli.Request()
	req.Method(`POST`)
	req.Path(`/api/v2/quota-scans`)
	req.JSON(&User)
	resp, err := req.Do()
	if err != nil {
		return err
	}
	if resp.Ok {
		return nil
	}

	return parseErrorResponse(resp)
}

func doJSON(req *gentleman.Request, dest interface{}) error {
	resp, err := req.Do()
	if err != nil {
		return err
	}

	if resp.Ok {
		return resp.JSON(&dest)
	}

	return parseErrorResponse(resp)
}

func parseErrorResponse(resp *gentleman.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
		// TODO: other error codes
	default:
		var apiErr apiError
		if err := resp.JSON(&apiErr); err != nil {
			return err
		}
		return &apiErr
	}
}
