package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

// Global variables
var (
	jwtMiddleware = middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(constants.JWT_SECRET),
		TokenLookup: "cookie:Authorization",
	})
)

// GitHub response model
type githubAuthResponse struct {
	AccessToken string `json:"access_token"`
}

// GitHub user model
type githubUserInfo struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

// Initialize authentication routes
func Init(auth *echo.Group) {
	auth.GET("/login", login)
	auth.GET("/loggedIn", loggedIn, jwtMiddleware)
}

func login(c echo.Context) error {
	// Get token from OAuth callback
	githubCode := c.QueryParam("code")
	logrus.WithField("code", githubCode).Debug("got code from github")

	// Get DB connection from middleware
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in login")
		return errors.New("Unable to get DB from context")
	}

	logrus.Debug("got database")

	// Create URL query object for requesting user access token from GitHub
	u := url.URL{Path: "https://github.com/login/oauth/access_token"}
	q := u.Query()
	q.Set("client_id", constants.GITHUB_CLIENT_ID)
	q.Set("client_secret", constants.GITHUB_CLIENT_SECRET)
	q.Set("code", githubCode)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url":   u,
			"error": err,
		}).Error("unable to create request to exchange github code")
		return err
	}

	logrus.Debug("made request to exchange github code")

	// Make request to GitHub with OAuth token for user access token
	req.Header.Add("Accept", "application/json")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"req":   req,
			"error": err,
		}).Error("unable to execute request to exchange github code")
		return err
	}
	defer resp.Body.Close()

	logrus.Debug("exchanged")

	// Parse access token response from GitHub
	parsedAuthResponse := githubAuthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&parsedAuthResponse)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"resp":  resp,
			"error": err,
		}).Error("unable to parse github auth response")
		return err
	}

	logrus.Debug("parsed auth response")

	// Create request to get user data using access token
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		logrus.WithError(err).Error("unable to create req to get github user id")
		return err
	}

	logrus.Debug("created req for user id")

	// Make request to GitHub for user data
	req.Header.Set("Authorization", "token "+parsedAuthResponse.AccessToken)
	resp, err = client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"req":   req,
			"error": err,
		}).Error("unable to execute req to get github user id")
		return err
	}
	defer resp.Body.Close()

	logrus.Debug("got response for user id")

	// Parse GitHub user data
	parsedUserInfo := githubUserInfo{}
	err = json.NewDecoder(resp.Body).Decode(&parsedUserInfo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"resp":  resp,
			"error": err,
		}).Error("unable to parse github user info response")
		return err
	}

	logrus.Debug("parsed github info")

	// Create user in DB
	user := models.User{GithubId: parsedUserInfo.ID}
	db.FirstOrCreate(&user, user)

	logrus.Debug("put user in db")

	// Create signed JWT with user's data. The JWT identifies the user to the app.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(constants.JWT_SECRET))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"token": token,
			"error": err,
		}).Error("Could not craft jwt")
		return err
	}

	logrus.Debug("got signed string")

	// Set authorization cookie for user
	cookie := new(echo.Cookie)
	cookie.SetName("Authorization")
	cookie.SetValue(t)
	cookie.SetHTTPOnly(true)
	cookie.SetSecure(true)
	cookie.SetPath("/")

	redir := "https://" + os.Getenv("DOMAIN")
	if os.Getenv("DEV") != "" {
		cookie.SetDomain("localhost")
		redir = redir + ":" + os.Getenv("PORT")
	}

	c.SetCookie(cookie)

	logrus.WithField("redir", redir).Debug("redirecting to")

	// Upon login, redirect user to homepage
	return c.Redirect(http.StatusFound, redir)
}

// there is a middleware in front to make sure the user is logged in
func loggedIn(c echo.Context) error {
	return nil
}
