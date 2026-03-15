package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

type testUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestOK(t *testing.T) {
	c, w := newTestContext()

	OK(c, testUser{Name: "alice", Age: 30})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response[testUser]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)
	assert.Equal(t, "alice", resp.Data.Name)
	assert.Equal(t, 30, resp.Data.Age)
}

func TestCreated(t *testing.T) {
	c, w := newTestContext()

	Created(c, testUser{Name: "bob", Age: 25})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response[testUser]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Status)
	assert.Equal(t, "bob", resp.Data.Name)
}

func TestOKList(t *testing.T) {
	c, w := newTestContext()

	items := []testUser{
		{Name: "alice", Age: 30},
		{Name: "bob", Age: 25},
	}
	OKList(c, items, 2)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response[ListData[testUser]]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Data.Total)
	assert.Len(t, resp.Data.Items, 2)
	assert.Equal(t, "alice", resp.Data.Items[0].Name)
}

func TestOKMessage(t *testing.T) {
	c, w := newTestContext()

	OKMessage(c, "deleted successfully")

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response[MessageData]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "deleted successfully", resp.Data.Message)
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(*gin.Context, string)
		status int
	}{
		{"BadRequest", BadRequest, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"Forbidden", Forbidden, http.StatusForbidden},
		{"NotFound", NotFound, http.StatusNotFound},
		{"Conflict", Conflict, http.StatusConflict},
		{"InternalError", InternalError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext()

			tt.fn(c, "test error")

			assert.Equal(t, tt.status, w.Code)

			var resp ErrorBody
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.status, resp.Status)
			assert.Equal(t, "test error", resp.Message)
		})
	}
}
