package controller

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/mobile-directing-system/mds-server/services/go/permission-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

// ControllerCreateUserSuite tests Controller.CreateUser.
type ControllerCreateUserSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	sampleUserID uuid.UUID
}

func (suite *ControllerCreateUserSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
}

func (suite *ControllerCreateUserSuite) TestStoreCreateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.sampleUserID).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerCreateUserSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("CreateUser", timeout, tx, suite.sampleUserID).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.CreateUser(timeout, tx, suite.sampleUserID)
		suite.Require().NoError(err, "should not fail")
	}()

	wait()
}

func TestController_CreateUser(t *testing.T) {
	suite.Run(t, new(ControllerCreateUserSuite))
}

// ControllerDeleteUserByIDSuite tests Controller.DeleteUserByID.
type ControllerDeleteUserByIDSuite struct {
	suite.Suite
	ctrl         *ControllerMock
	sampleUserID uuid.UUID
}

func (suite *ControllerDeleteUserByIDSuite) SetupTest() {
	suite.ctrl = NewMockController()
	suite.sampleUserID = testutil.NewUUIDV4()
}

func (suite *ControllerDeleteUserByIDSuite) TestPermissionUpdateFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, []store.Permission{}).
		Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestNotifyFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, []store.Permission{}).
		Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", timeout, tx, suite.sampleUserID, []store.Permission{}).
		Return(errors.New("sad life"))
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestDeleteUserFail() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, []store.Permission{}).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).Return(errors.New("sad life"))
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", timeout, tx, suite.sampleUserID, []store.Permission{}).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.Error(err, "should fail")
	}()

	wait()
}

func (suite *ControllerDeleteUserByIDSuite) TestOK() {
	timeout, cancel, wait := testutil.NewTimeout(suite, timeout)
	tx := &testutil.DBTx{}
	suite.ctrl.DB.Tx = []*testutil.DBTx{{}}
	suite.ctrl.Store.On("UpdatePermissionsByUser", timeout, tx, suite.sampleUserID, []store.Permission{}).
		Return(nil)
	suite.ctrl.Store.On("DeleteUserByID", timeout, tx, suite.sampleUserID).Return(nil)
	defer suite.ctrl.Store.AssertExpectations(suite.T())
	suite.ctrl.Notifier.On("NotifyPermissionsUpdated", timeout, tx, suite.sampleUserID, []store.Permission{}).Return(nil)
	defer suite.ctrl.Notifier.AssertExpectations(suite.T())

	go func() {
		defer cancel()
		err := suite.ctrl.Ctrl.DeleteUserByID(timeout, tx, suite.sampleUserID)
		suite.NoError(err, "should not fail")
	}()

	wait()
}

func TestController_DeleteUserByID(t *testing.T) {
	suite.Run(t, new(ControllerDeleteUserByIDSuite))
}
