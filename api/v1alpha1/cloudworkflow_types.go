/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkfloStatus int32

const (
	WorkfloStatus_UNSPECIFIED WorkfloStatus = 0
	WorkfloStatus_Pending     WorkfloStatus = 1
	WorkfloStatus_Success     WorkfloStatus = 2
	WorkfloStatus_Failure     WorkfloStatus = 3
)

type WorkflowStepType int32

const (
	WorkflowStepType_Customizable  WorkflowStepType = 0
	WorkflowStepType_CodePull      WorkflowStepType = 1
	WorkflowStepType_ImageRepoAuth WorkflowStepType = 2
	WorkflowStepType_Build         WorkflowStepType = 3
	WorkflowStepType_Deploy        WorkflowStepType = 4
)

// WorkflowStepType to string
func (w WorkflowStepType) String() string {
	switch w {
	case WorkflowStepType_Customizable:
		return "Customizable"
	case WorkflowStepType_CodePull:
		return "CodePull"
	case WorkflowStepType_ImageRepoAuth:
		return "ImageRepoAuth"
	case WorkflowStepType_Build:
		return "Build"
	case WorkflowStepType_Deploy:
		return "Deploy"
	default:
		return ""
	}
}

type WorkflowType int32

const (
	WorkflowType_UNSPECIFIED               WorkflowType = 0
	WorkflowType_ContinuousIntegrationType WorkflowType = 1
	WorkflowType_ContinuousDeploymentType  WorkflowType = 2
)

// WorkflowType to string
func (w WorkflowType) String() string {
	switch w {
	case WorkflowType_ContinuousIntegrationType:
		return "ContinuousIntegrationType"
	case WorkflowType_ContinuousDeploymentType:
		return "ContinuousDeploymentType"
	default:
		return ""
	}
}

type WorkflowTask struct {
	Id          int64         `json:"id,omitempty" gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	WorkflowId  int64         `json:"workflow_id,omitempty" gorm:"column:workflow_id;default:0;NOT NULL;index:idx_workflow_id"`
	StepId      int64         `json:"step_id,omitempty" gorm:"column:step_id;default:0;NOT NULL"`
	Name        string        `json:"name,omitempty" gorm:"column:name;default:'';NOT NULL"`
	Order       int32         `json:"order,omitempty" gorm:"column:order;default:0;NOT NULL"`
	TaskCommand string        `json:"task_command,omitempty" gorm:"column:task_command;default:'';NOT NULL"`
	Description string        `json:"description,omitempty" gorm:"column:description;default:'';NOT NULL"`
	Status      WorkfloStatus `json:"status,omitempty" gorm:"column:status;default:0;NOT NULL"`
	Log         string        `json:"log,omitempty" gorm:"-"`
}

type WorkflowStep struct {
	Id               int64            `json:"id,omitempty" gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	WorkflowId       int64            `json:"workflow_id,omitempty" gorm:"column:workflow_id;default:0;NOT NULL;index:idx_workflow_id"`
	Order            int32            `json:"order,omitempty" gorm:"column:order;default:0;NOT NULL"`
	Name             string           `json:"name,omitempty" gorm:"column:name;default:'';NOT NULL"`
	Description      string           `json:"description,omitempty" gorm:"column:description;default:'';NOT NULL"`
	Image            string           `json:"image,omitempty" gorm:"column:image;default:'';NOT NULL"`
	WorkflowStepType WorkflowStepType `json:"workflow_step_type,omitempty" gorm:"column:workflow_step_type;default:0;NOT NULL"`
	WorkflowTasks    []*WorkflowTask  `json:"workflow_tasks,omitempty" gorm:"-"`
}

type Workflow struct {
	Id            int64           `json:"id,omitempty" gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	Name          string          `json:"name,omitempty" gorm:"column:name;default:'';NOT NULL"`
	Namespace     string          `json:"namespace,omitempty" gorm:"column:namespace;default:'';NOT NULL"`
	Lables        string          `json:"lables,omitempty" gorm:"column:lables;default:'';NOT NULL"`
	Env           string          `json:"env,omitempty" gorm:"column:env;default:'';NOT NULL"`
	StorageClass  string          `json:"storage_class,omitempty" gorm:"column:storage_class;default:'';NOT NULL"`
	Type          WorkflowType    `json:"type,omitempty" gorm:"column:type;default:0;NOT NULL"`
	Description   string          `json:"description,omitempty" gorm:"column:description;default:'';NOT NULL"`
	ServiceId     int64           `json:"service_id,omitempty" gorm:"column:service_id;default:0;NOT NULL;index:idx_service_id"`
	WorkflowSteps []*WorkflowStep `json:"workflow_steps,omitempty" gorm:"-"`
}

func (w *Workflow) GetFirstStep() *WorkflowStep {
	if len(w.WorkflowSteps) == 0 {
		return nil
	}
	var firstStep *WorkflowStep
	for _, step := range w.WorkflowSteps {
		if firstStep == nil || step.Order < firstStep.Order {
			firstStep = step
		}
	}
	return firstStep
}

func (w *Workflow) GetNextStep(s *WorkflowStep) *WorkflowStep {
	if len(w.WorkflowSteps) == 0 {
		return nil
	}
	var nextStep *WorkflowStep
	for _, step := range w.WorkflowSteps {
		if step.Order > s.Order {
			if nextStep == nil || step.Order < nextStep.Order {
				nextStep = step
			}
		}
	}
	return nextStep
}

func (s *WorkflowStep) GetFirstTask() *WorkflowTask {
	if len(s.WorkflowTasks) == 0 {
		return nil
	}
	var firstTask *WorkflowTask
	for _, task := range s.WorkflowTasks {
		if firstTask == nil || task.Order < firstTask.Order {
			firstTask = task
		}
	}
	return firstTask
}

func (s *WorkflowStep) GetNextTask(t *WorkflowTask) *WorkflowTask {
	if len(s.WorkflowTasks) == 0 {
		return nil
	}
	var nextTask *WorkflowTask
	for _, task := range s.WorkflowTasks {
		if task.Order > t.Order {
			if nextTask == nil || task.Order < nextTask.Order {
				nextTask = task
			}
		}
	}
	return nextTask
}

// Get all tasks by order step task
func (w *Workflow) GetAllTasksByOrder() []*WorkflowTask {
	steps := make([]*WorkflowStep, len(w.WorkflowSteps))
	copy(steps, w.WorkflowSteps)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	var allTasks []*WorkflowTask
	for _, step := range steps {
		tasks := make([]*WorkflowTask, len(step.WorkflowTasks))
		copy(tasks, step.WorkflowTasks)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Order < tasks[j].Order
		})
		allTasks = append(allTasks, tasks...)
	}
	return allTasks
}

func (w *Workflow) GetTask(taskName string) *WorkflowTask {
	for _, step := range w.WorkflowSteps {
		for _, task := range step.WorkflowTasks {
			if task.Name == taskName {
				return task
			}
		}
	}
	return nil
}

func (w *Workflow) GetStorageName() string {
	return w.Name + "-storage"
}

func (w *Workflow) GetWorkdir() string {
	return "/app"
}

func (w *Workflow) GetWorkdirName() string {
	return "app"
}

type WorkflowStatus int32

const (
	WorkflowStatusPending  WorkflowStatus = 0
	WorkflowStatusRunning                 = 1
	WorkflowStatusSucceede                = 2
	WorkflowStatusFaile                   = 3
)

// CloudWorkflowStatus defines the observed state of CloudWorkflow.
type CloudWorkflowStatus struct {
	Status  WorkflowStatus `json:"status,omitempty"`
	Message string         `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudWorkflow is the Schema for the cloudworkflows API.
type CloudWorkflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Workflow            `json:"spec,omitempty"`
	Status CloudWorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudWorkflowList contains a list of CloudWorkflow.
type CloudWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudWorkflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudWorkflow{}, &CloudWorkflowList{})
}
