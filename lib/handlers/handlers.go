package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/satori/go.uuid"
)

func SetupRoutes(mux *chi.Mux, db Backend, log zerolog.Logger) *chi.Mux {
	timeHandler := TimeHandler{
		Db:  db,
		Log: log,
	}

	mux.Route("/time", func(r chi.Router) {
		r.Post("/", timeHandler.CreateTime)
		r.Get("/{timeId}", timeHandler.GetTime)
		r.Put("/{timeId}", timeHandler.ChangeTime)
		r.Delete("/{timeId}", timeHandler.DeleteTime)
	})

	return mux
}

func (t *TimeHandler) CreateTime(w http.ResponseWriter, r *http.Request) {
	// default start time
	timeStr := "12:00 PM"

	if r.ContentLength > 0 {
		defer r.Body.Close()
		bdy, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		newTime := NewTimeRequest{}
		err = json.Unmarshal(bdy, &newTime)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !validTimeFormat(newTime.InitialTime) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		timeStr = newTime.InitialTime
	}

	id := uuid.NewV4().String()
	err := t.Db.SetTimeId(id, timeStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(NewTime{
		TimeId:      id,
		CurrentTime: timeStr,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(resp)
	if err != nil {
		t.Log.Debug().
			Err(err).
			Msg("failure during write response")
	}
}

func (t *TimeHandler) GetTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "timeId")

	if _, err := uuid.FromString(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	val, err := t.Db.GetTimeId(id)
	if err != nil {
		if t.Db.NotFoundErrCheck(err) {
			t.Log.Debug().Err(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res CurrentTime
	res.CurrentTime = val
	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(resp)
	if err != nil {
		t.Log.Debug().
			Err(err).
			Msg("failure during write response")
	}
}

func (t *TimeHandler) ChangeTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "timeId")

	if _, err := uuid.FromString(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	bdy, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	timeChange := ChangeTimeRequest{}
	err = json.Unmarshal(bdy, &timeChange)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	current, err := t.Db.GetTimeId(id)
	if err != nil {
		if t.Db.NotFoundErrCheck(err) {
			t.Log.Debug().Err(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newTime := calculateTime(current, timeChange.AddMinutes)
	err = t.Db.SetTimeId(id, newTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res CurrentTime
	res.CurrentTime = newTime

	resp, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(resp)
	if err != nil {
		t.Log.Debug().
			Err(err).
			Msg("failure during write response")
	}
}

func (t *TimeHandler) DeleteTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "timeId")
	if _, err := uuid.FromString(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := t.Db.DeleteTimeId(id)
	if err != nil {
		if t.Db.NotFoundErrCheck(err) {
			t.Log.Debug().Err(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
