// Copyright © 2017 Wikia Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"net/http"

	"github.com/go-playground/lars"
	"github.com/google/go-github/github"
	"github.com/harnash/emperor/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

func handleMain(ctx lars.Context) {
	resp := ctx.Response()

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.Write([]byte(htmlIndex))
}

func handleGitHubLogin(ctx lars.Context) {
	url := getOAuthConfig().AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	resp := ctx.Response()
	http.Redirect(resp.Writer(), ctx.Request(), url, http.StatusTemporaryRedirect)
}

func handleGitHubCallback(ctx lars.Context) {
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

func handleLogged(ctx lars.Context) {
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

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an API server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting server")

		l := lars.New()
		l.Use(server.LoggingAndRecovery)

		l.Get("/", handleMain)
		l.Get("/login", handleGitHubLogin)
		l.Get("/logged", handleLogged)
		l.Get("/github_oauth_cb", handleGitHubCallback)

		http.ListenAndServe(":8080", l.Serve())
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
