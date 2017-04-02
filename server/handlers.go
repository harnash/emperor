package server

import (
	"fmt"
	"net/http"

	"github.com/go-playground/lars"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

const htmlIndex = `<html><body>
Log in with <a href="/login">GitHub</a>
</body></html>`

var (
	oauthStateString = "CHANGEME"
)

func getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     viper.GetString("githubclientid"),
		ClientSecret: viper.GetString("githubsecret"),
		RedirectURL:  "http://localhost:8080/github_oauth_cb",
		Scopes: []string{
			"user:email",
			"read:org",
		},
		Endpoint: githuboauth.Endpoint,
	}
}

func HandleMain(ctx lars.Context) {
	resp := ctx.Response()

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.Write([]byte(htmlIndex))
}

func HandleGitHubLogin(ctx lars.Context) {
	url := getOAuthConfig().AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	resp := ctx.Response()
	http.Redirect(resp.Writer(), ctx.Request(), url, http.StatusTemporaryRedirect)
}

func HandleGitHubCallback(ctx lars.Context) {
	req := ctx.Request()
	resp := ctx.Response()

	state := req.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(resp.Writer(), req, "/", http.StatusTemporaryRedirect)
		return
	}

	code := req.FormValue("code")
	oauthConf := getOAuthConfig()
	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(resp.Writer(), req, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := oauthConf.Client(ctx, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get("")
	if err != nil {
		log.WithError(err).Error("client.Users.Get() faled with")
		http.Redirect(resp.Writer(), req, "/", http.StatusTemporaryRedirect)
		return
	}
	log.WithField("user", *user.Login).Info("Logged in as GitHub user")
	http.Redirect(resp.Writer(), req, fmt.Sprintf("/logged?token=%s", token.AccessToken), http.StatusTemporaryRedirect)
}

func HandleLogged(ctx lars.Context) {
	resp := ctx.Response()
	req := ctx.Request()
	code := req.FormValue("token")

	log.Debug(code)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: code},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	orgs, _, err := client.Organizations.List("", nil)

	if err != nil {
		log.WithError(err).Error("Error fetching organizations")
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp.WriteString(fmt.Sprintf("organizations: %s", orgs))
}
