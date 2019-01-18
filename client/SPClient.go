package spclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const SharePointServicePrincipal = "00000003-0000-0ff1-ce00-000000000000"

type SPClient struct {
	clientId      string
	clientSecret  string
	siteUrl       string
	httpClient    *http.Client
	token         *AuthToken
	authUrlResult *AuthUrlResult
	authUrl       string
	principal     string
	hostname      string
	realm         string
}

func NewSPClient(siteUrl, clientId, clientSecret string) *SPClient {
	return &SPClient{
		siteUrl:    siteUrl,
		httpClient: &http.Client{},
		token:      nil,
	}
}

func GetRealm(siteUrl string) string {
	r, err := http.NewRequest("POST", siteUrl+"/vti_bin/client.svc", nil)
	r.Header.Add("Authorization", "Bearer ")
	client := &http.Client{}

	resp, err := client.Do(r)
	if err == nil {
		var t = resp.Header.Get("www-authenticate")
		idx := strings.Index(t, "Bearer realm=\"")

		return t[idx+14 : idx+50]
	}
	return ""
}

func (client *SPClient) GetRealm() string {
	r, err := http.NewRequest("POST", client.siteUrl+"/vti_bin/client.svc", nil)
	r.Header.Add("Authorization", "Bearer ")
	resp, err := client.httpClient.Do(r)
	if err == nil {
		var t = resp.Header.Get("www-authenticate")
		idx := strings.Index(t, "Bearer realm=\"")

		return t[idx+14 : idx+50]
	} else {
		fmt.Println(err)
	}
	return ""
}

func GetAuthUrl(realm string) string {
	uu := "https://accounts.accesscontrol.windows.net/metadata/json/1?realm=" + realm
	resp, _ := http.Get(uu)
	arr, _ := ioutil.ReadAll(resp.Body)
	var authUrl AuthUrlResult
	_ = json.Unmarshal(arr, &authUrl)
	l := len(authUrl.EndPoints)
	for i := 0; i < l; i++ {
		if authUrl.EndPoints[i].Protocol == "OAuth2" {
			return authUrl.EndPoints[i].Location
		}
	}
	return ""
}

func GetFormattedPrincipal(principal, hostname, realm string) string {
	var r = principal
	if hostname != "" {
		r += "/" + hostname
	}
	r += "@" + realm
	return r
}

func (client *SPClient) GetFormattedPrincipal() string {
	var r = client.principal
	if client.hostname != "" {
		r += "/" + client.hostname
	}
	r += "@" + client.realm
	return r
}

func (client *SPClient) GetAuthUrl() string {
	uu := "https://accounts.accesscontrol.windows.net/metadata/json/1?realm=" + client.realm
	resp, _ := client.httpClient.Get(uu)
	arr, _ := ioutil.ReadAll(resp.Body)
	var authUrl AuthUrlResult
	_ = json.Unmarshal(arr, &authUrl)
	l := len(authUrl.EndPoints)
	client.authUrlResult = &authUrl
	for i := 0; i < l; i++ {
		if authUrl.EndPoints[i].Protocol == "OAuth2" {
			client.authUrl = authUrl.EndPoints[i].Location
			return client.authUrl
		}
	}
	return ""
}

func (client *SPClient) hasTokenExpired() bool {
	tt, err := strconv.ParseInt(client.token.ExpiresOn, 10, 64)
	if err == nil {
		if tt >= 10000000000 {
			tt = tt / 1000
		}
		return time.Now().Before(time.Unix(tt, 0))
	}
	return true
}

func (client *SPClient) GetAddInOnlyAccessToken() *AuthToken {
	if client.token != nil {
		if !client.hasTokenExpired() {
			return client.token
		}
	}
	sUrl, _ := url.Parse(client.siteUrl)
	resUrl := GetFormattedPrincipal(SharePointServicePrincipal, sUrl.Hostname(), client.realm)
	fmtClientId := GetFormattedPrincipal(client.clientId, "", client.realm)
	authUrl := GetAuthUrl(client.realm)
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_id", fmtClientId)
	v.Set("client_secret", client.clientSecret)
	v.Set("resource", resUrl)

	r, _ := http.NewRequest("POST", authUrl, bytes.NewReader([]byte(v.Encode())))
	resp, _ := client.httpClient.Do(r)
	arr, _ := ioutil.ReadAll(resp.Body)
	var token AuthToken
	_ = json.Unmarshal(arr, &token)
	client.token = &token
	return client.token
}

func (client *SPClient) Get(destUrl string) []byte {
	token := client.GetAddInOnlyAccessToken()
	req, _ := http.NewRequest("GET", destUrl, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json;odata=verbose;charset=utf-8")
	resp, err := client.httpClient.Do(req)
	if err == nil {
		arr, e := ioutil.ReadAll(resp.Body)
		if e == nil {
			return arr
		}
	}
	return nil
}

func (client *SPClient) Delete(destUrl string) []byte {
	token := client.GetAddInOnlyAccessToken()
	req, _ := http.NewRequest("DELETE", destUrl, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json;odata=verbose;charset=utf-8")
	resp, err := client.httpClient.Do(req)
	if err == nil {
		arr, e := ioutil.ReadAll(resp.Body)
		if e == nil {
			return arr
		}
	}
	return nil
}

func (client *SPClient) PostJson(destUrl string, body []byte) []byte {
	token := client.GetAddInOnlyAccessToken()
	reader := bytes.NewReader(body)
	req, _ := http.NewRequest("POST", destUrl, reader)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json;odata=verbose;charset=utf-8")
	resp, err := client.httpClient.Do(req)
	if err == nil {
		arr, e := ioutil.ReadAll(resp.Body)
		if e == nil {
			return arr
		}
	}
	return nil
}

type AuthUrlEndPoint struct {
	Protocol string `json:"protocol"`
	Location string `json:"location"`
}

type AuthUrlResult struct {
	EndPoints []AuthUrlEndPoint `json:"endpoints"`
}

type AuthToken struct {
	TokenType   string `json:"token_type"`
	NotBefore   string `json:"not_before"`
	ExpiresIn   string `json:"expires_in"`
	ExpiresOn   string `json:"expires_on"`
	Resource    string `json:"resource"`
	AccessToken string `json:"access_token"`
}

func GetAddInOnlyAccessToken(siteUrl, realm, clientId, clientSecret string) AuthToken {
	sUrl, _ := url.Parse(siteUrl)
	resUrl := GetFormattedPrincipal(SharePointServicePrincipal, sUrl.Hostname(), realm)
	fmtClientId := GetFormattedPrincipal(clientId, "", realm)
	authUrl := GetAuthUrl(realm)
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_id", fmtClientId)
	v.Set("client_secret", clientSecret)
	v.Set("resource", resUrl)

	r, _ := http.NewRequest("POST", authUrl, bytes.NewReader([]byte(v.Encode())))
	client := http.Client{}
	resp, _ := client.Do(r)
	arr, _ := ioutil.ReadAll(resp.Body)
	var token AuthToken
	_ = json.Unmarshal(arr, &token)
	return token
}

func main() {
	fmt.Println("Good one")
	clientId := "9e2bc1cd-7a31-4cdb-8c6e-8a05a4adcfcc"
	clientSecret := "ohZCPSXpGHmrGM8xN0G3y5+xOPWkbw1HasxriVkUwBM="
	u := "https://jlrglobal.sharepoint.com/sites/jlrway"
	uu, _ := url.Parse(u)

	realm := GetRealm(u)
	prin := GetFormattedPrincipal(SharePointServicePrincipal, uu.Hostname(), realm)

	fmt.Println(realm)
	fmt.Println(prin)
	fmt.Println(GetAuthUrl(realm))
	token := GetAddInOnlyAccessToken(u, realm, clientId, clientSecret)
	fmt.Println(token.AccessToken)

	rr, _ := http.NewRequest("GET", "https://jlrglobal.sharepoint.com/sites/jlrway/en-gb/_api/web/lists/getByTitle('JLR Emails')", nil)
	rr.Header.Add("Authorization", "Bearer "+token.AccessToken)
	rr.Header.Add("accept", "application/json")
	rr.Header.Add("content-type", "application/json;odata=verbose;charset=utf-8")
	client := http.Client{}
	resp, err := client.Do(rr)
	if err == nil {
		arr, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(arr))
	} else {
		fmt.Println(err)
	}

}
