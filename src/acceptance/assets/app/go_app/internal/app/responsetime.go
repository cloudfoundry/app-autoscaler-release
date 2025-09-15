package app

import (
	"net/http"
	"strconv"
	"time"
)

//counterfeiter:generate . TimeWaster
type TimeWaster interface {
	Sleep(sleepTime time.Duration)
}

type Sleeper struct{}

var _ TimeWaster = Sleeper{}

func ResponseTimeTests(mux *http.ServeMux, timeWaster TimeWaster) {
	mux.HandleFunc("GET /responsetime/slow/{delayInMS}", func(w http.ResponseWriter, r *http.Request) {
		var milliseconds int64
		var err error
		if milliseconds, err = strconv.ParseInt(r.PathValue("delayInMS"), 10, 64); err != nil {
			Error(w, http.StatusBadRequest, "invalid milliseconds: %s", err.Error())
			return
		}
		duration := time.Duration(milliseconds) * time.Millisecond
		timeWaster.Sleep(duration)
		writeJSON(w, http.StatusOK, JSONResponse{"duration": duration.String()})
	})

	mux.HandleFunc("GET /responsetime/fast", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, JSONResponse{"fast": true})
	})
}

func (Sleeper) Sleep(sleepTime time.Duration) {
	time.Sleep(sleepTime)
}
