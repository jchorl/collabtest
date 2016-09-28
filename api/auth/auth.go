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

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

type githubAuthResponse struct {
	AccessToken string `json:"access_token"`
}

type githubUserInfo struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

func Init(auth *echo.Group) {
	auth.GET("/login", login)
}

func login(c echo.Context) error {
	githubCode := c.QueryParam("code")
	logrus.WithField("code", githubCode).Debug("got code from github")

	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in login")
		return errors.New("Unable to get DB from context")
	}

	logrus.Debug("got database")

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

	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		logrus.WithError(err).Error("unable to create req to get github user id")
		return err
	}

	logrus.Debug("created req for user id")

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

	user := models.User{GithubId: parsedUserInfo.ID}
	db.FirstOrCreate(&user, user)

	logrus.Debug("put user in db")

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

	cookie := new(echo.Cookie)
	cookie.SetName("Authorization")
	cookie.SetValue(t)
	c.SetCookie(cookie)

	redir := "https://" + os.Getenv("DOMAIN")
	if os.Getenv("DEV") != "" {
		redir = redir + ":" + os.Getenv("PORT")
	}

	logrus.WithField("redir", redir).Debug("redirecting to")

	return c.Redirect(http.StatusFound, redir)
}
