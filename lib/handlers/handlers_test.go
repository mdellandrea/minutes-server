package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
)

type testBackend struct{}

func (b *testBackend) SetTimeId(id, val string) error      { return nil }
func (b *testBackend) GetTimeId(id string) (string, error) { return "12:00 PM", nil }
func (b *testBackend) DeleteTimeId(id string) error        { return nil }
func (b *testBackend) NotFoundErrCheck(error) bool         { return false }

type testBackendFail struct{}

func (b *testBackendFail) SetTimeId(id, val string) error      { return fmt.Errorf("Err") }
func (b *testBackendFail) GetTimeId(id string) (string, error) { return "", fmt.Errorf("Err") }
func (b *testBackendFail) DeleteTimeId(id string) error        { return fmt.Errorf("Err") }
func (b *testBackendFail) NotFoundErrCheck(error) bool         { return false }

type testBackendNotFound struct{}

func (b *testBackendNotFound) SetTimeId(id, val string) error      { return fmt.Errorf("Err") }
func (b *testBackendNotFound) GetTimeId(id string) (string, error) { return "", fmt.Errorf("Err") }
func (b *testBackendNotFound) DeleteTimeId(id string) error        { return fmt.Errorf("Err") }
func (b *testBackendNotFound) NotFoundErrCheck(error) bool         { return true }

var testTimeHandler = TimeHandler{
	Db: &testBackend{},
}

var failingTestTimeHandler = TimeHandler{
	Db: &testBackendFail{},
}

var notFoundTestTimeHandler = TimeHandler{
	Db: &testBackendNotFound{},
}

func TestNewRouter(t *testing.T) {
	mux := chi.NewMux()
	logger := zerolog.New(os.Stderr)
	testRtr := SetupRoutes(mux, &testBackend{}, logger)
	x := testRtr.Routes()
	if len(x) > 1 {
		t.Errorf("root pattern length: got <%d> want <%d>", len(x), 1)
	}

	expectedRoot := "/time/*"
	actualRoot := x[0].Pattern
	if x[0].Pattern != expectedRoot {
		t.Errorf("root pattern: got <%s> want <%s>", actualRoot, expectedRoot)
	}

	routes1 := x[0].SubRoutes.Routes()
	noParamsRoutes := routes1[0]
	expectedRoutes1 := "/"
	if noParamsRoutes.Pattern != expectedRoutes1 {
		t.Errorf("no configured route: %s", expectedRoutes1)
	}
	postHandler := noParamsRoutes.Handlers["POST"]
	if postHandler == nil {
		t.Error("no configured POST handler")
	}

	routes2 := x[0].SubRoutes.Routes()
	paramsRoutes := routes2[1]
	expectedRoutes2 := "/{timeId}"
	if paramsRoutes.Pattern != expectedRoutes2 {
		t.Errorf("no configured route: %s", expectedRoutes2)
	}
	getHandler := paramsRoutes.Handlers["GET"]
	if getHandler == nil {
		t.Error("no configured GET handler")
	}
	putHandler := paramsRoutes.Handlers["PUT"]
	if putHandler == nil {
		t.Error("no configured PUT handler")
	}
	deleteHandler := paramsRoutes.Handlers["DELETE"]
	if deleteHandler == nil {
		t.Error("no configured DELETE handler")
	}
}

func TestCreateTimeHandler(t *testing.T) {
	t.Run("No Request Body - Success", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/time", nil)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("TestCreateTimeHandler No Request Body - Success - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusOK)
		}

		responseHeaders := rr.Header()
		contentType := responseHeaders.Get("Content-Type")
		if contentType != "application/json; charset=UTF-8" {
			t.Errorf("TestCreateTimeHandler No Request Body - Success - Content-Type Header: got <%s> want <%s>", contentType, "application/json; charset=UTF-8")
		}
	})

	t.Run("Request Body - Success", func(t *testing.T) {
		b := strings.NewReader(`{"initialTime":"03:33 PM"}`)
		req, err := http.NewRequest("POST", "/time", b)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("TestCreateTimeHandler Request Body - Success - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusOK)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json; charset=UTF-8" {
			t.Errorf("TestCreateTimeHandler Request Body - Success - Content-Type Header: got <%s> want <%s>", contentType, "application/json; charset=UTF-8")
		}

		bdy, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error(err)
		}
		var tgt NewTime
		err = json.Unmarshal(bdy, &tgt)
		if err != nil {
			t.Errorf("TestCreateTimeHandler Request Body - Success - JSON Response Unmarshal failed: <%s>", err)
		}

		if _, err := uuid.FromString(tgt.TimeId); err != nil {
			t.Errorf("TestCreateTimeHandler Request Body - Success - JSON Response timeId invalid format: got <%s> want <%s>", tgt.TimeId, "uuidV4")
		}

		if tgt.CurrentTime != "03:33 PM" {
			t.Errorf("TestCreateTimeHandler Request Body - Success - JSON Response currentTime invalid: got <%s> want <%s>", tgt.CurrentTime, "03:33 PM")
		}
	})

	t.Run("Request Body - Invalid Request Format", func(t *testing.T) {
		b := strings.NewReader(`{"spongeBob":"squarePants"}`)
		req, err := http.NewRequest("POST", "/time", b)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestCreateTimeHandler Request Body - Invalid Time Format - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("Request Body - Invalid Time Format", func(t *testing.T) {
		b := strings.NewReader(`{"initialTime":"13:33 PM"}`)
		req, err := http.NewRequest("POST", "/time", b)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestCreateTimeHandler Request Body - Invalid Time Format - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("Request Body - Malformed JSON Failure", func(t *testing.T) {
		b := strings.NewReader(`{"initialTime":12X}`)
		req, err := http.NewRequest("POST", "/time", b)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestCreateTimeHandler Request Body - Malformed JSON Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		b := strings.NewReader(`{"initialTime":"03:33 PM"}`)
		req, err := http.NewRequest("POST", "/time", b)
		if err != nil {
			t.Error(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(failingTestTimeHandler.CreateTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("TestCreateTimeHandler Request Body - DB Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusInternalServerError)
		}
	})
}

func TestGetTimeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.GetTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("TestGetTimeHandler - Success - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusOK)
		}

		expected := []byte(`{"currentTime":"12:00 PM"}`)
		if !bytes.Equal(rr.Body.Bytes(), expected) {
			t.Errorf("TestGetTimeHandler - Success - Response Body: got <%v> want <%v>", rr.Body.Bytes(), expected)
		}
	})

	t.Run("Invalid timeId Format", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", "12345")
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.GetTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestGetTimeHandler - Invalid timeId Format- Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("timeId Not Found", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(notFoundTestTimeHandler.GetTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("TestGetTimeHandler - timeId Not Found- Response Status Code: got <%d> want <%d>", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		r, err := http.NewRequest("GET", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(failingTestTimeHandler.GetTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("TestGetTimeHandler - DB Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusInternalServerError)
		}
	})
}

func TestChangeTimeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b := strings.NewReader(`{"addMinutes":10}`)
		r, err := http.NewRequest("PUT", "/time", b)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("TestChangeTimeHandler - Success - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusOK)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json; charset=UTF-8" {
			t.Errorf("TestChangeTimeHandler - Success - Content-Type Header: got <%s> want <%s>", contentType, "application/json; charset=UTF-8")
		}

		bdy, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error(err)
		}
		var tgt CurrentTime
		err = json.Unmarshal(bdy, &tgt)
		if err != nil {
			t.Errorf("TestChangeTimeHandler - Success - JSON Response Unmarshal failed: <%s>", err)
		}

		// Start time is 12:00 PM. Request is for 10
		if tgt.CurrentTime != "12:10 PM" {
			t.Errorf("TestChangeTimeHandler - Success - JSON Response currentTime invalid: got <%s> want <%s>", tgt.CurrentTime, "12:10 PM")
		}
	})

	t.Run("Invalid timeId Format", func(t *testing.T) {
		r, err := http.NewRequest("PUT", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", "ZZZZ")
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestChangeTimeHandler - Invalid timeId Format - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("timeId Not Found", func(t *testing.T) {
		b := strings.NewReader(`{"addMinutes":12345}`)
		r, err := http.NewRequest("PUT", "/time", b)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(notFoundTestTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("TestChangeTimeHandler - timeId Not Found- Response Status Code: got <%d> want <%d>", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("No Request Body", func(t *testing.T) {
		r, err := http.NewRequest("PUT", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestChangeTimeHandler - No Request Body - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("Request Body - Malformed JSON Failure", func(t *testing.T) {
		b := strings.NewReader(`{"addMinutes":12X}`)
		r, err := http.NewRequest("PUT", "/time", b)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestChangeTimeHandler - Request Body - Malformed JSON Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("Request Body - Invalid Request Format", func(t *testing.T) {
		b := strings.NewReader(`{"addMinutes":"derp"}`)
		r, err := http.NewRequest("PUT", "/time", b)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestChangeTimeHandler - Request Body - Invalid Request Format - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		b := strings.NewReader(`{"addMinutes":12345}`)
		r, err := http.NewRequest("PUT", "/time", b)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(failingTestTimeHandler.ChangeTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("TestChangeTimeHandler - DB Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusInternalServerError)
		}
	})
}

func TestDeleteTimeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, err := http.NewRequest("DELETE", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		ctx := chi.NewRouteContext()
		u2 := uuid.NewV4()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.DeleteTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Errorf("TestDeleteTimeHandler - Success - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusNoContent)
		}
	})

	t.Run("Invalid timeId Format", func(t *testing.T) {
		r, err := http.NewRequest("DELETE", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", "12345")
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(testTimeHandler.DeleteTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("TestDeleteTimeHandler - Invalid timeId Format - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("timeId Not Found", func(t *testing.T) {
		r, err := http.NewRequest("DELETE", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(notFoundTestTimeHandler.DeleteTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("TestDeleteTimeHandler - timeId Not Found- Response Status Code: got <%d> want <%d>", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		r, err := http.NewRequest("DELETE", "/time", nil)
		if err != nil {
			t.Error(err)
		}
		u2 := uuid.NewV4()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("timeId", u2.String())
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(failingTestTimeHandler.DeleteTime)

		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("TestDeleteTimeHandler - DB Failure - Response Status Code: got <%d> want <%d>", rr.Code, http.StatusInternalServerError)
		}
	})
}
