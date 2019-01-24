// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package restrictedshell implements session shell plugin with restricted commands.
package restrictedshell

import (
	"fmt"
	"testing"

	"github.com/aws/amazon-ssm-agent/agent/appconfig"
	"github.com/aws/amazon-ssm-agent/agent/context"
	"github.com/aws/amazon-ssm-agent/agent/contracts"
	"github.com/aws/amazon-ssm-agent/agent/framework/processor/executer/iohandler"
	"github.com/aws/amazon-ssm-agent/agent/framework/processor/executer/iohandler/mock"
	"github.com/aws/amazon-ssm-agent/agent/log"
	mgsContracts "github.com/aws/amazon-ssm-agent/agent/session/contracts"
	dataChannelMock "github.com/aws/amazon-ssm-agent/agent/session/datachannel/mocks"
	"github.com/aws/amazon-ssm-agent/agent/session/plugins/sessionplugin/mocks"
	"github.com/aws/amazon-ssm-agent/agent/session/plugins/shell"
	"github.com/aws/amazon-ssm-agent/agent/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RestrictedShellTestSuite struct {
	suite.Suite
	mockContext     *context.Mock
	mockLog         log.T
	mockCancelFlag  *task.MockCancelFlag
	mockDataChannel *dataChannelMock.IDataChannel
	mockIohandler   *iohandlermocks.MockIOHandler
	plugin          *RestrictedShellPlugin
}

func (suite *RestrictedShellTestSuite) SetupTest() {
	mockContext := context.NewMockDefault()
	mockCancelFlag := &task.MockCancelFlag{}
	mockDataChannel := &dataChannelMock.IDataChannel{}
	mockIohandler := new(iohandlermocks.MockIOHandler)
	mockLog := log.NewMockLog()

	suite.mockContext = mockContext
	suite.mockLog = mockLog
	suite.mockCancelFlag = mockCancelFlag
	suite.mockDataChannel = mockDataChannel
	suite.mockIohandler = mockIohandler
	suite.plugin = &RestrictedShellPlugin{}
}

//Execute the test suite
func TestRestrictedShellTestSuite(t *testing.T) {
	suite.Run(t, new(RestrictedShellTestSuite))
}

// Testing Name
func (suite *RestrictedShellTestSuite) TestName() {
	rst := suite.plugin.name()
	assert.Equal(suite.T(), rst, appconfig.PluginNameRestrictedStandardStream)
}

// Testing Execute
func (suite *RestrictedShellTestSuite) TestExecuteWhenCancelFlagIsShutDown() {
	suite.mockCancelFlag.On("ShutDown").Return(true)
	suite.mockIohandler.On("MarkAsShutdown").Return(nil)
	suite.plugin.shell, _ = shell.NewPlugin()

	suite.plugin.Execute(suite.mockContext,
		contracts.Configuration{Commands: "test"},
		suite.mockCancelFlag,
		suite.mockIohandler,
		suite.mockDataChannel)

	suite.mockCancelFlag.AssertExpectations(suite.T())
	suite.mockIohandler.AssertExpectations(suite.T())
}

// Testing Execute
func (suite *RestrictedShellTestSuite) TestExecuteWhenCancelFlagIsCancelled() {
	suite.mockCancelFlag.On("Canceled").Return(true)
	suite.mockCancelFlag.On("ShutDown").Return(false)
	suite.mockIohandler.On("MarkAsCancelled").Return(nil)
	suite.plugin.shell, _ = shell.NewPlugin()

	suite.plugin.Execute(suite.mockContext,
		contracts.Configuration{Commands: "test"},
		suite.mockCancelFlag,
		suite.mockIohandler,
		suite.mockDataChannel)

	suite.mockCancelFlag.AssertExpectations(suite.T())
	suite.mockIohandler.AssertExpectations(suite.T())
}

// Testing Execute
func (suite *RestrictedShellTestSuite) TestExecute() {
	newIOHandler := iohandler.NewDefaultIOHandler(suite.mockLog, contracts.IOConfiguration{})
	mockShellPlugin := new(mocks.ISessionPlugin)
	mockShellPlugin.On("Execute", suite.mockContext, mock.Anything, suite.mockCancelFlag, newIOHandler, suite.mockDataChannel).Return()
	suite.plugin.shell = mockShellPlugin

	suite.plugin.Execute(suite.mockContext,
		contracts.Configuration{Commands: "test"},
		suite.mockCancelFlag,
		newIOHandler,
		suite.mockDataChannel)

	mockShellPlugin.AssertExpectations(suite.T())
	assert.Equal(suite.T(), 0, newIOHandler.GetExitCode())
}

// Testing Execute
func (suite *RestrictedShellTestSuite) TestExecuteWithoutCommands() {
	suite.mockIohandler.On("SetExitCode", 1).Return(nil)
	suite.mockIohandler.On("SetStatus", contracts.ResultStatusFailed).Return()

	sessionPluginResultOutput := mgsContracts.SessionPluginResultOutput{}
	sessionPluginResultOutput.Output = fmt.Sprintf("Commands cannot be empty for session type %s", suite.plugin.name())
	suite.mockIohandler.On("SetOutput", sessionPluginResultOutput).Return()

	suite.plugin.Execute(suite.mockContext,
		contracts.Configuration{},
		suite.mockCancelFlag,
		suite.mockIohandler,
		suite.mockDataChannel)

	suite.mockIohandler.AssertExpectations(suite.T())
}

// Testing InputStreamMessageHandler
func (suite *RestrictedShellTestSuite) TestInputStreamMessageHandler() {
	mockShellPlugin := new(mocks.ISessionPlugin)
	mockShellPlugin.On("InputStreamMessageHandler", suite.mockLog, mock.Anything).Return(nil)
	suite.plugin.shell = mockShellPlugin

	err := suite.plugin.InputStreamMessageHandler(suite.mockLog, mgsContracts.AgentMessage{})

	mockShellPlugin.AssertExpectations(suite.T())
	assert.Nil(suite.T(), err)
}