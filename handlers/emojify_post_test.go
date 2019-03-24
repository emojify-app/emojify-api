package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockEmojifyer emojify.ClientMock
var mockCache cache.ClientMock

func resetCacheMock() {
	mockCache.ExpectedCalls = make([]*mock.Call, 0)
}

func setupEmojiHandler() (*httptest.ResponseRecorder, *http.Request, *EmojifyPost) {
	mockEmojifyer = emojify.ClientMock{}
	mockCache = cache.ClientMock{}

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	logger, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	h := NewEmojifyPost(logger, &mockEmojifyer, &mockCache)

	return rw, r, h
}

func TestReturnsBadRequestIfBodyLessThan8(t *testing.T) {
	rw, r, h := setupEmojiHandler()

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, " is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestReturnsInvalidURLIfBodyNotURL(t *testing.T) {
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("httsddfdfdf/cc")))

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "httsddfdfdf/cc is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestReturns302IfImageIsCached(t *testing.T) {
	rw, r, h := setupEmojiHandler()

	u, _ := url.Parse(fileURL)
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(u.String())))

	resetCacheMock()
	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: true},
		nil,
	)

	pretty.Println(mockCache.ExpectedCalls)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotModified, rw.Code)
}

func TestCallsEmojifyIfNotCached(t *testing.T) {

}