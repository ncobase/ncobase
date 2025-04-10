package service

import (
	nec "github.com/ncobase/ncore/ext/core"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/repository"
)

// Service represents the workflow service
type Service struct {
	Process       ProcessServiceInterface
	Node          NodeServiceInterface
	Task          TaskServiceInterface
	Template      TemplateServiceInterface
	Business      BusinessServiceInterface
	History       HistoryServiceInterface
	ProcessDesign ProcessDesignServiceInterface
	Delegation    DelegationServiceInterface
	Rule          RuleServiceInterface
}

// New creates a new workflow service
func New(d *data.Data, em nec.ManagerInterface) *Service {
	// Create repositories
	repo := repository.New(d)

	// Create services
	return &Service{
		Process:       NewProcessService(repo, em),
		Node:          NewNodeService(repo, em),
		Task:          NewTaskService(repo, em),
		Template:      NewTemplateService(repo, em),
		Business:      NewBusinessService(repo, em),
		History:       NewHistoryService(repo, em),
		ProcessDesign: NewProcessDesignService(repo, em),
		Delegation:    NewDelegationService(repo, em),
		Rule:          NewRuleService(repo, em),
	}
}

// GetProcess returns process service
func (s *Service) GetProcess() ProcessServiceInterface {
	return s.Process
}

// GetNode returns node service
func (s *Service) GetNode() NodeServiceInterface {
	return s.Node
}

// GetTask returns task service
func (s *Service) GetTask() TaskServiceInterface {
	return s.Task
}

// GetTemplate returns template service
func (s *Service) GetTemplate() TemplateServiceInterface {
	return s.Template
}

// GetBusiness returns business service
func (s *Service) GetBusiness() BusinessServiceInterface {
	return s.Business
}

// GetHistory returns history service
func (s *Service) GetHistory() HistoryServiceInterface {
	return s.History
}

// GetProcessDesign returns process design service
func (s *Service) GetProcessDesign() ProcessDesignServiceInterface {
	return s.ProcessDesign
}

// GetDelegation returns delegation service
func (s *Service) GetDelegation() DelegationServiceInterface {
	return s.Delegation
}

// GetRule returns rule service
func (s *Service) GetRule() RuleServiceInterface {
	return s.Rule
}
