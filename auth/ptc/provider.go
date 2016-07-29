package ptc

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	authorizeURL = "https://sso.pokemon.com/sso/oauth2.0/accessToken"
	loginURL     = "https://sso.pokemon.com/sso/login?service=https://sso.pokemon.com/sso/oauth2.0/callbackAuthorize"

	redirectURI = "https://www.nianticlabs.com/pokemon/error"

	clientSecret = "w8ScCUXJQc6kXKw8FiOhd8Fixzht18Dq3PEVkUCP5ZPxtgyWsbTvWHFLm2wNY0JR"
	clientID     = "mobile-app_pokemon-go"
)

const providerString = "ptc"

// LoginRequest - type to handle Auth Request
type LoginRequest struct {
	Lt        string   `json:"lt"`
	Execution string   `json:"execution"`
	Errors    []string `json:"errors,omitempty"`
}

// Provider - PTC auth main provider type
type Provider struct {
	username string
	password string
	ticket   string
	http     *http.Client
}

// HTTPResponses - handling async request
type HTTPResponses struct {
	url      string
	response *http.Response
	rawBody  []byte
	err      error
}

// NewProvider - Create Pokemon Trainer Club auth provider instance
func NewProvider(username, password string) *Provider {
	options := &cookiejar.Options{}
	jar, _ := cookiejar.New(options)

	httpClient := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("Use the last error")
		},
	}

	return &Provider{
		http:     httpClient,
		username: username,
		password: password,
	}
}

// GetProviderString - return PTC as provider type
func (p *Provider) GetProviderString() string {
	return providerString
}

// GetAccessToken - return PTC access token
func (p *Provider) GetAccessToken() string {
	return p.ticket
}

// Login - PTC login method
func (p *Provider) Login() (string, error) {
	ch := make(chan *HTTPResponses, 1)

	go p.checkLoginProcess(ch)
	checkLoginResp := <-ch

	if checkLoginResp.err != nil {
		return loginError("Could not start the login process, website might be down")
	}

	go p.processLogin(ch, checkLoginResp)
	processLoginResp := <-ch

	if processLoginResp.err != nil {
		return loginError(processLoginResp.err.Error())
	}

	go p.processTicket(ch, processLoginResp)
	close(ch)

	return p.ticket, nil
}

func (p *Provider) checkLoginProcess(ch chan<- *HTTPResponses) {

	req, _ := http.NewRequest("GET", loginURL, nil)
	req.Header.Set("User-Agent", "niantic")

	resp, err := p.http.Do(req)
	if err != nil {
		ch <- &HTTPResponses{loginURL, nil, nil, err}
		return
	}

	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		ch <- &HTTPResponses{loginURL, nil, nil, err2}
		return
	}

	ch <- &HTTPResponses{loginURL, resp, body, nil}
}

func (p *Provider) processLogin(ch chan<- *HTTPResponses, resp *HTTPResponses) {
	var loginRespBody LoginRequest
	json.Unmarshal(resp.rawBody, &loginRespBody)

	loginForm := url.Values{}
	loginForm.Set("lt", loginRespBody.Lt)
	loginForm.Set("execution", loginRespBody.Execution)
	loginForm.Set("_eventId", "submit")
	loginForm.Set("username", p.username)
	loginForm.Set("password", p.password)

	loginFormData := strings.NewReader(loginForm.Encode())

	req, _ := http.NewRequest("POST", loginURL, loginFormData)
	req.Header.Set("User-Agent", "niantic")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respLogin, err := p.http.Do(req)
	if _, ok := err.(*url.Error); !ok {
		defer respLogin.Body.Close()
		rawBody, _ := ioutil.ReadAll(respLogin.Body)
		var respBody LoginRequest
		json.Unmarshal(rawBody, &respBody)

		var errorVar error

		if len(respBody.Errors) > 0 {
			_, errorVar = loginError(respBody.Errors[0])
		} else {
			_, errorVar = loginError("Could not request authorization")
		}

		ch <- &HTTPResponses{loginURL, respLogin, rawBody, errorVar}
		return
	}

	ch <- &HTTPResponses{loginURL, respLogin, nil, nil}
	return
}

func (p *Provider) processTicket(ch chan<- *HTTPResponses, resp *HTTPResponses) {
	location, _ := url.Parse(resp.response.Header.Get("Location"))
	ticket := location.Query().Get("ticket")

	authForm := url.Values{}
	authForm.Set("client_id", clientID)
	authForm.Set("redirect_uri", redirectURI)
	authForm.Set("client_secret", clientSecret)
	authForm.Set("grant_type", "refresh_token")
	authForm.Set("code", ticket)

	authFormData := strings.NewReader(authForm.Encode())

	req, _ := http.NewRequest("POST", authorizeURL, authFormData)
	req.Header.Set("User-Agent", "niantic")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respAuth, err := p.http.Do(req)
	if err != nil {
		ch <- &HTTPResponses{authorizeURL, nil, nil, err}
		return
	}

	defer respAuth.Body.Close()
	b, _ := ioutil.ReadAll(respAuth.Body)
	query, _ := url.ParseQuery(string(b))
	p.ticket = query.Get("access_token")
}
