package repository

import "ncobase/workflow/data"

// Repository represents the collection of all repositories
type Repository interface {
	GetProcess() ProcessRepositoryInterface
	GetNode() NodeRepositoryInterface
	GetTask() TaskRepositoryInterface
	GetTemplate() TemplateRepositoryInterface
	GetBusiness() BusinessRepositoryInterface
	GetHistory() HistoryRepositoryInterface
	GetProcessDesign() ProcessDesignRepositoryInterface
	GetDelegation() DelegationRepositoryInterface
	GetRule() RuleRepositoryInterface
}

// repository implements Repository interface
type repository struct {
	process       ProcessRepositoryInterface
	node          NodeRepositoryInterface
	task          TaskRepositoryInterface
	template      TemplateRepositoryInterface
	business      BusinessRepositoryInterface
	history       HistoryRepositoryInterface
	processDesign ProcessDesignRepositoryInterface
	delegation    DelegationRepositoryInterface
	rule          RuleRepositoryInterface
}

// New creates a new repository
func New(d *data.Data) Repository {
	return &repository{
		process:       NewProcessRepository(d),
		node:          NewNodeRepository(d),
		task:          NewTaskRepository(d),
		template:      NewTemplateRepository(d),
		business:      NewBusinessRepository(d),
		history:       NewHistoryRepository(d),
		processDesign: NewProcessDesignRepository(d),
		delegation:    NewDelegationRepository(d),
		rule:          NewRuleRepository(d),
	}
}

// GetProcess returns process repository
func (r *repository) GetProcess() ProcessRepositoryInterface {
	return r.process
}

// GetNode returns node repository
func (r *repository) GetNode() NodeRepositoryInterface {
	return r.node
}

// GetTask returns task repository
func (r *repository) GetTask() TaskRepositoryInterface {
	return r.task
}

// GetTemplate returns template repository
func (r *repository) GetTemplate() TemplateRepositoryInterface {
	return r.template
}

// GetBusiness returns business repository
func (r *repository) GetBusiness() BusinessRepositoryInterface {
	return r.business
}

// GetHistory returns history repository
func (r *repository) GetHistory() HistoryRepositoryInterface {
	return r.history
}

// GetProcessDesign returns process design repository
func (r *repository) GetProcessDesign() ProcessDesignRepositoryInterface {
	return r.processDesign
}

// GetDelegation returns delegation repository
func (r *repository) GetDelegation() DelegationRepositoryInterface {
	return r.delegation
}

// GetRule returns rule repository
func (r *repository) GetRule() RuleRepositoryInterface {
	return r.rule
}
