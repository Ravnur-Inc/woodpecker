// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	scheduler_mocks "go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCancel(t *testing.T) {
	repo := &model.Repo{ID: 1, FullName: "org/repo"}
	user := &model.User{ID: 1}
	pipeline := &model.Pipeline{ID: 2, RepoID: 1, Status: model.StatusRunning}

	runningWorkflow := &model.Workflow{ID: 3, PipelineID: 2, State: model.StatusRunning}
	pendingWorkflow := &model.Workflow{
		ID: 4, PipelineID: 2, State: model.StatusPending,
		Children: []*model.Step{{ID: 5, State: model.StatusPending}},
	}

	mockStore := store_mocks.NewMockStore(t)
	mockScheduler := scheduler_mocks.NewMockScheduler(t)
	mockForge := forge_mocks.NewMockForge(t)

	origScheduler := server.Config.Services.Scheduler
	server.Config.Services.Scheduler = mockScheduler
	t.Cleanup(func() { server.Config.Services.Scheduler = origScheduler })

	var calls []string

	mockStore.On("WorkflowGetTree", mock.Anything).
		Return([]*model.Workflow{runningWorkflow, pendingWorkflow}, nil).Once()
	mockStore.On("UpdatePipeline", mock.Anything).Run(func(args mock.Arguments) {
		assert.Equal(t, model.StatusKilled, args.Get(0).(*model.Pipeline).Status)
		calls = append(calls, "UpdatePipeline")
	}).Return(nil).Once()
	mockScheduler.On("CancelWorkflows", mock.Anything, []string{"3", "4"}).Run(func(mock.Arguments) {
		calls = append(calls, "CancelWorkflows")
	}).Return(nil).Once()
	mockStore.On("WorkflowUpdate", mock.Anything).Return(nil).Once()
	mockStore.On("StepUpdate", mock.Anything).Return(nil).Once()

	skippedWorkflow := *pendingWorkflow
	skippedWorkflow.State = model.StatusSkipped
	mockStore.On("WorkflowGetTree", mock.Anything).
		Return([]*model.Workflow{runningWorkflow, &skippedWorkflow}, nil).Once()
	mockForge.On("Status", mock.Anything, user, repo, mock.Anything, runningWorkflow).Return(nil).Once()
	mockForge.On("Status", mock.Anything, user, repo, mock.Anything, &skippedWorkflow).Return(nil).Once()
	mockScheduler.On("PublishPipelineEvent", mock.Anything, repo, mock.Anything).Return(nil).Once()

	err := Cancel(t.Context(), mockForge, mockStore, repo, user, pipeline, &model.CancelInfo{CanceledByUser: "someone"})
	require.NoError(t, err)

	// the pipeline must be marked as killed before the workflows are evicted from the queue,
	// so agents that check the store see the cancellation
	assert.Equal(t, []string{"UpdatePipeline", "CancelWorkflows"}, calls)
}
