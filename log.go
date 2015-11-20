// Walkhub
// Copyright (C) 2015 Pronovix
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package walkhub

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/nbio/hitch"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/tamasd/ab"
	"github.com/tamasd/hitch-session"
)

//go:generate abt --output=loggen.go --generate-crud-update=false --generate-crud-delete=false --generate-service-struct=false --generate-service-struct-name=LogService --generate-service-list=false --generate-service-get=false --generate-service-post=false --generate-service-put=false --generate-service-delete=false --generate-service-patch=false entity Log

type Log struct {
	UUID    string    `dbtype:"uuid" dbdefault:"uuid_generate_v4()" json:"uuid"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Created time.Time `dbdefault:"now()" json:"created"`
}

func DBLog(db ab.DB, logtype, message string) error {
	l := &Log{
		Type:    logtype,
		Message: message,
		Created: time.Now(),
	}

	return l.Insert(db)
}

type helpCenterOpenedLog struct {
	URL string `json:"url"`
}

type walkthroughPlayedLog struct {
	UUID         string `json:"uuid"`
	ErrorMessage string `json:"errorMessage"`
	EmbedOrigin  string `json:"embedOrigin"`
}

type walkthroughPageVisitedLog struct {
	UUID        string `json:"uuid"`
	EmbedOrigin string `json:"embedOrigin"`
}

func getLogUserID(r *http.Request) string {
	db := ab.GetDB(r)
	userid := r.RemoteAddr
	uid := session.GetSession(r)["uid"]
	if uid != "" {
		user, err := LoadUser(db, uid)
		if err != nil {
			log.Println(err)
		} else {
			userid = user.Mail
		}
	}

	return userid
}

type LogService struct {
	BaseURL string
}

func (s *LogService) Register(h *hitch.Hitch) error {
	walkthroughPlayed := prometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "walkhub",
		Subsystem: "metrics",
		Name:      "walkthrough_played",
		Help:      "Number of walkthrough plays",
	}, []string{"uuid", "embedorigin"})

	walkthroughVisited := prometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "walkhub",
		Subsystem: "metrics",
		Name:      "walkthrough_visited",
		Help:      "Number of walkthrough visits",
	}, []string{"uuid", "embedorigin"})

	h.Post("/api/log/helpcenteropened", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := helpCenterOpenedLog{}
		ab.MustDecode(r, &l)

		db := ab.GetDB(r)
		userid := getLogUserID(r)

		message := fmt.Sprintf("%s has opened the help center on %s", userid, l.URL)
		ab.MaybeFail(r, http.StatusInternalServerError, DBLog(db, "helpcenteropened", message))
	}))

	h.Post("/api/log/walkthroughplayed", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := walkthroughPlayedLog{}
		ab.MustDecode(r, &l)

		db := ab.GetDB(r)
		userid := getLogUserID(r)
		wt, err := LoadActualRevision(db, l.UUID)
		ab.MaybeFail(r, http.StatusBadRequest, err)
		if wt == nil {
			ab.Fail(r, http.StatusNotFound, nil)
		}

		message := ""

		embedPart := ""
		if l.EmbedOrigin != "" {
			embedPart = "from the help center on " + l.EmbedOrigin + " "
		}

		wturl := s.BaseURL + "walkthrough/" + wt.UUID

		if l.ErrorMessage == "" {
			message = fmt.Sprintf("%s has played the walkthrough %s<%s|%s>", userid, embedPart, wturl, wt.Name)
		} else {
			message = fmt.Sprintf("%s has failed to play the walkthrough %s<%s|%s> with the error message %s", userid, embedPart, wturl, wt.Name, l.ErrorMessage)
		}

		ab.MaybeFail(r, http.StatusInternalServerError, DBLog(db, "walkthroughplayed", message))

		walkthroughPlayed.
			With(metrics.Field{Key: "uuid", Value: l.UUID}).
			With(metrics.Field{Key: "embedorigin", Value: l.EmbedOrigin}).
			Add(1)
	}))

	h.Post("/api/log/walkthroughpagevisited", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := walkthroughPageVisitedLog{}
		ab.MustDecode(r, &l)

		db := ab.GetDB(r)
		userid := getLogUserID(r)
		wt, err := LoadActualRevision(db, l.UUID)
		ab.MaybeFail(r, http.StatusBadRequest, err)
		if wt == nil {
			ab.Fail(r, http.StatusNotFound, nil)
		}

		embedPart := ""
		if l.EmbedOrigin != "" {
			embedPart = "embedded on " + l.EmbedOrigin + " "
		}

		wturl := s.BaseURL + "walkthrough/" + wt.UUID

		message := fmt.Sprintf("%s has visited the walkthrough page %s<%s|%s>", userid, embedPart, wturl, wt.Name)

		ab.MaybeFail(r, http.StatusInternalServerError, DBLog(db, "walkthroughvisited", message))

		walkthroughVisited.
			With(metrics.Field{Key: "uuid", Value: l.UUID}).
			With(metrics.Field{Key: "embedorigin", Value: l.EmbedOrigin}).
			Add(1)
	}))

	return nil
}
