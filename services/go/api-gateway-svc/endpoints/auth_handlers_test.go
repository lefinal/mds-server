package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/api-gateway-svc/controller"
	"github.com/mobile-directing-system/mds-server/services/go/shared/auth"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

// handleLoginStoreMock mocks handleLoginStore.
type handleLoginStoreMock struct {
	mock.Mock
}

func (m *handleLoginStoreMock) Login(ctx context.Context, username string, pass string,
	requestMetadata controller.AuthRequestMetadata) (string, bool, error) {
	args := m.Called(ctx, username, pass, requestMetadata)
	return args.String(0), args.Bool(1), args.Error(2)
}

// handleLoginSuite tests handleLogin.
type handleLoginSuite struct {
	suite.Suite
	s             *handleLoginStoreMock
	r             *gin.Engine
	sampleRequest loginPayload
}

func (suite *handleLoginSuite) SetupTest() {
	suite.s = &handleLoginStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/login", handleLogin(zap.NewNop(), suite.s))
	suite.sampleRequest = loginPayload{
		Username: "sweat",
		Pass:     "bind",
	}
}

func (suite *handleLoginSuite) TestInvalidBody() {
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   strings.NewReader("{invalid"),
	})
	suite.Equal(http.StatusBadRequest, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestLoginFail() {
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return("", false, errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestBadLogin() {
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return("", false, nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Equal(http.StatusUnauthorized, rr.Code, "should return correct code")
}

func (suite *handleLoginSuite) TestOK() {
	token := "feed"
	suite.s.On("Login", mock.Anything, suite.sampleRequest.Username, suite.sampleRequest.Pass, mock.Anything).
		Return(token, true, nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/login",
		Body:   bytes.NewReader(testutil.MarshalJSONMust(suite.sampleRequest)),
	})
	suite.Require().Equal(http.StatusOK, rr.Code, "should return correct code")
	var got loginResponse
	suite.Require().NoError(json.NewDecoder(rr.Body).Decode(&got), "should return valid response")
	suite.Equal(loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, got, "should return correct response")
}

func Test_handleLogin(t *testing.T) {
	suite.Run(t, new(handleLoginSuite))
}

func Test_extractAuthRequestMetadataFromRequest(t *testing.T) {
	host := "sheet"
	userAgent := "yes"
	remoteAddr := "ocean"
	req := &http.Request{
		Host:       host,
		RemoteAddr: remoteAddr,
		Header: http.Header{
			"User-Agent": []string{userAgent},
		},
	}
	got := extractAuthRequestMetadataFromRequest(req)
	assert.Equal(t, controller.AuthRequestMetadata{
		Host:       host,
		UserAgent:  userAgent,
		RemoteAddr: remoteAddr,
	}, got, "should extract correct metadata")
}

// handleLogoutStoreMock mocks handleLogoutStore.
type handleLogoutStoreMock struct {
	mock.Mock
}

func (m *handleLogoutStoreMock) Logout(ctx context.Context, publicToken string,
	requestMetadata controller.AuthRequestMetadata) error {
	return m.Called(ctx, publicToken, requestMetadata).Error(0)
}

// handleLogoutSuite tests handleLogout.
type handleLogoutSuite struct {
	suite.Suite
	s              *handleLogoutStoreMock
	r              *gin.Engine
	sampleToken    auth.Token
	sampleTokenStr string
}

func (suite *handleLogoutSuite) SetupTest() {
	suite.s = &handleLogoutStoreMock{}
	suite.r = testutil.NewGinEngine()
	suite.r.POST("/logout", handleLogout(zap.NewNop(), suite.s))
	suite.sampleToken = auth.Token{
		UserID: uuid.New(),
	}
	var err error
	suite.sampleTokenStr, err = auth.GenJWTToken(suite.sampleToken, "")
	if err != nil {
		panic(err)
	}
}

func (suite *handleLogoutSuite) TestLogoutFail() {
	suite.s.On("Logout", mock.Anything, suite.sampleTokenStr, mock.Anything).Return(errors.New("sad life"))
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/logout",
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusInternalServerError, rr.Code, "should return correct code")
}

func (suite *handleLogoutSuite) TestOK() {
	suite.s.On("Logout", mock.Anything, suite.sampleTokenStr, mock.Anything).Return(nil)
	defer suite.s.AssertExpectations(suite.T())
	rr := testutil.DoHTTPRequestMust(testutil.HTTPRequestProps{
		Server: suite.r,
		Method: http.MethodPost,
		URL:    "/logout",
		Token:  suite.sampleToken,
	})
	suite.Equal(http.StatusOK, rr.Code, "should return correct code")
}

func Test_handleLogout(t *testing.T) {
	suite.Run(t, new(handleLogoutSuite))
}
