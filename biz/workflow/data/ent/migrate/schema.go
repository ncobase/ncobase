// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// NcseFlowBusinessColumns holds the columns for the "ncse_flow_business" table.
	NcseFlowBusinessColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "code", Type: field.TypeString, Nullable: true, Comment: "code"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "form_code", Type: field.TypeString, Comment: "Form type code"},
		{Name: "form_version", Type: field.TypeString, Nullable: true, Comment: "Form version number"},
		{Name: "form_config", Type: field.TypeJSON, Nullable: true, Comment: "Form configuration"},
		{Name: "form_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Form permission settings"},
		{Name: "field_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Field level permissions"},
		{Name: "process_id", Type: field.TypeString, Comment: "Process instance ID"},
		{Name: "template_id", Type: field.TypeString, Comment: "Process template ID"},
		{Name: "business_key", Type: field.TypeString, Comment: "Business document ID"},
		{Name: "origin_data", Type: field.TypeJSON, Comment: "Original form data"},
		{Name: "current_data", Type: field.TypeJSON, Comment: "Current form data"},
		{Name: "change_logs", Type: field.TypeJSON, Nullable: true, Comment: "Data change history"},
		{Name: "last_modified", Type: field.TypeInt64, Nullable: true, Comment: "Last modification time"},
		{Name: "last_modifier", Type: field.TypeString, Nullable: true, Comment: "Last modifier"},
		{Name: "operation_logs", Type: field.TypeJSON, Nullable: true, Comment: "Operation logs"},
		{Name: "flow_status", Type: field.TypeString, Nullable: true, Comment: "Flow status"},
		{Name: "flow_variables", Type: field.TypeJSON, Nullable: true, Comment: "Flow variables"},
		{Name: "is_draft", Type: field.TypeBool, Comment: "Whether is draft", Default: false},
		{Name: "is_terminated", Type: field.TypeBool, Comment: "Whether is terminated", Default: false},
		{Name: "is_suspended", Type: field.TypeBool, Comment: "Whether is suspended", Default: false},
		{Name: "suspend_reason", Type: field.TypeString, Nullable: true, Comment: "Suspension reason"},
		{Name: "business_tags", Type: field.TypeJSON, Nullable: true, Comment: "Business tags"},
		{Name: "module_code", Type: field.TypeString, Comment: "Module code"},
		{Name: "category", Type: field.TypeString, Nullable: true, Comment: "Category"},
		{Name: "viewers", Type: field.TypeJSON, Nullable: true, Comment: "Users with view permission"},
		{Name: "editors", Type: field.TypeJSON, Nullable: true, Comment: "Users with edit permission"},
		{Name: "permission_configs", Type: field.TypeJSON, Nullable: true, Comment: "Permission configurations"},
		{Name: "role_configs", Type: field.TypeJSON, Nullable: true, Comment: "Role configurations"},
		{Name: "visible_range", Type: field.TypeJSON, Nullable: true, Comment: "Visibility range"},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
	}
	// NcseFlowBusinessTable holds the schema information for the "ncse_flow_business" table.
	NcseFlowBusinessTable = &schema.Table{
		Name:       "ncse_flow_business",
		Columns:    NcseFlowBusinessColumns,
		PrimaryKey: []*schema.Column{NcseFlowBusinessColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "business_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowBusinessColumns[0]},
			},
			{
				Name:    "business_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowBusinessColumns[32]},
			},
			{
				Name:    "business_business_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowBusinessColumns[10]},
			},
			{
				Name:    "business_process_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowBusinessColumns[8]},
			},
			{
				Name:    "business_module_code_form_code",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowBusinessColumns[24], NcseFlowBusinessColumns[3]},
			},
			{
				Name:    "business_flow_status",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowBusinessColumns[17]},
			},
		},
	}
	// NcseFlowDelegationColumns holds the columns for the "ncse_flow_delegation" table.
	NcseFlowDelegationColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "delegator_id", Type: field.TypeString, Comment: "User ID who delegates"},
		{Name: "delegatee_id", Type: field.TypeString, Comment: "User ID to delegate to"},
		{Name: "template_id", Type: field.TypeString, Nullable: true, Comment: "Template ID if specific"},
		{Name: "node_type", Type: field.TypeString, Nullable: true, Comment: "Node type if specific"},
		{Name: "conditions", Type: field.TypeJSON, Nullable: true, Comment: "Delegation conditions"},
		{Name: "start_time", Type: field.TypeInt64, Comment: "Delegation start time"},
		{Name: "end_time", Type: field.TypeInt64, Comment: "Delegation end time"},
		{Name: "is_enabled", Type: field.TypeBool, Comment: "Whether delegation is enabled", Default: true},
	}
	// NcseFlowDelegationTable holds the schema information for the "ncse_flow_delegation" table.
	NcseFlowDelegationTable = &schema.Table{
		Name:       "ncse_flow_delegation",
		Columns:    NcseFlowDelegationColumns,
		PrimaryKey: []*schema.Column{NcseFlowDelegationColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "delegation_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowDelegationColumns[0]},
			},
			{
				Name:    "delegation_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowDelegationColumns[3]},
			},
			{
				Name:    "delegation_delegator_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowDelegationColumns[8]},
			},
			{
				Name:    "delegation_template_id_node_type",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowDelegationColumns[10], NcseFlowDelegationColumns[11]},
			},
		},
	}
	// NcseFlowHistoryColumns holds the columns for the "ncse_flow_history" table.
	NcseFlowHistoryColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "process_id", Type: field.TypeString, Comment: "Process instance ID"},
		{Name: "template_id", Type: field.TypeString, Comment: "Process template ID"},
		{Name: "business_key", Type: field.TypeString, Comment: "Business document ID"},
		{Name: "node_key", Type: field.TypeString, Unique: true, Comment: "Unique identifier for the node"},
		{Name: "node_type", Type: field.TypeString, Comment: "Node type"},
		{Name: "node_config", Type: field.TypeJSON, Nullable: true, Comment: "Node configuration"},
		{Name: "node_rules", Type: field.TypeJSON, Nullable: true, Comment: "Node rules"},
		{Name: "node_events", Type: field.TypeJSON, Nullable: true, Comment: "Node events"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "node_name", Type: field.TypeString, Comment: "Node name"},
		{Name: "operator", Type: field.TypeString, Comment: "Operation user"},
		{Name: "operator_dept", Type: field.TypeString, Nullable: true, Comment: "Operator's department"},
		{Name: "task_id", Type: field.TypeString, Nullable: true, Comment: "Task ID"},
		{Name: "variables", Type: field.TypeJSON, Comment: "Task variables"},
		{Name: "form_data", Type: field.TypeJSON, Nullable: true, Comment: "Form data"},
		{Name: "action", Type: field.TypeString, Comment: "Operation action"},
		{Name: "comment", Type: field.TypeString, Nullable: true, Comment: "Operation comment"},
		{Name: "details", Type: field.TypeJSON, Nullable: true, Comment: "Detailed information"},
	}
	// NcseFlowHistoryTable holds the schema information for the "ncse_flow_history" table.
	NcseFlowHistoryTable = &schema.Table{
		Name:       "ncse_flow_history",
		Columns:    NcseFlowHistoryColumns,
		PrimaryKey: []*schema.Column{NcseFlowHistoryColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "history_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowHistoryColumns[0]},
			},
			{
				Name:    "history_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowHistoryColumns[10]},
			},
			{
				Name:    "history_process_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowHistoryColumns[2]},
			},
			{
				Name:    "history_operator",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowHistoryColumns[16]},
			},
			{
				Name:    "history_action",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowHistoryColumns[21]},
			},
		},
	}
	// NcseFlowNodeColumns holds the columns for the "ncse_flow_node" table.
	NcseFlowNodeColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "node_key", Type: field.TypeString, Unique: true, Comment: "Unique identifier for the node"},
		{Name: "node_type", Type: field.TypeString, Comment: "Node type"},
		{Name: "node_config", Type: field.TypeJSON, Nullable: true, Comment: "Node configuration"},
		{Name: "node_rules", Type: field.TypeJSON, Nullable: true, Comment: "Node rules"},
		{Name: "node_events", Type: field.TypeJSON, Nullable: true, Comment: "Node events"},
		{Name: "form_code", Type: field.TypeString, Comment: "Form type code"},
		{Name: "form_version", Type: field.TypeString, Nullable: true, Comment: "Form version number"},
		{Name: "form_config", Type: field.TypeJSON, Nullable: true, Comment: "Form configuration"},
		{Name: "form_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Form permission settings"},
		{Name: "field_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Field level permissions"},
		{Name: "assignees", Type: field.TypeJSON, Comment: "Task assignees"},
		{Name: "candidates", Type: field.TypeJSON, Comment: "Candidate assignees"},
		{Name: "delegated_from", Type: field.TypeString, Nullable: true, Comment: "Delegated from user"},
		{Name: "delegated_reason", Type: field.TypeString, Nullable: true, Comment: "Delegation reason"},
		{Name: "is_delegated", Type: field.TypeBool, Comment: "Whether task is delegated", Default: false},
		{Name: "is_transferred", Type: field.TypeBool, Comment: "Whether task is transferred", Default: false},
		{Name: "allow_cancel", Type: field.TypeBool, Comment: "Allow cancellation", Default: true},
		{Name: "allow_urge", Type: field.TypeBool, Comment: "Allow urging", Default: true},
		{Name: "allow_delegate", Type: field.TypeBool, Comment: "Allow delegation", Default: true},
		{Name: "allow_transfer", Type: field.TypeBool, Comment: "Allow transfer", Default: true},
		{Name: "is_draft_enabled", Type: field.TypeBool, Comment: "Whether draft is enabled", Default: true},
		{Name: "is_auto_start", Type: field.TypeBool, Comment: "Whether auto start is enabled", Default: false},
		{Name: "strict_mode", Type: field.TypeBool, Comment: "Enable strict mode", Default: false},
		{Name: "start_time", Type: field.TypeInt64, Comment: "Start time"},
		{Name: "end_time", Type: field.TypeInt64, Nullable: true, Comment: "End time"},
		{Name: "due_time", Type: field.TypeInt64, Nullable: true, Comment: "Due time"},
		{Name: "duration", Type: field.TypeInt, Nullable: true, Comment: "Duration in seconds"},
		{Name: "priority", Type: field.TypeInt, Comment: "Priority level", Default: 0},
		{Name: "is_timeout", Type: field.TypeBool, Comment: "Whether timed out", Default: false},
		{Name: "reminder_count", Type: field.TypeInt, Comment: "Number of reminders sent", Default: 0},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "process_id", Type: field.TypeString, Comment: "Process ID"},
		{Name: "permissions", Type: field.TypeJSON, Comment: "Permission configs"},
		{Name: "prev_nodes", Type: field.TypeJSON, Nullable: true, Comment: "Previous nodes"},
		{Name: "next_nodes", Type: field.TypeJSON, Nullable: true, Comment: "Next nodes"},
		{Name: "parallel_nodes", Type: field.TypeJSON, Nullable: true, Comment: "Parallel nodes"},
		{Name: "branch_nodes", Type: field.TypeJSON, Nullable: true, Comment: "Branch nodes"},
		{Name: "conditions", Type: field.TypeJSON, Nullable: true, Comment: "Transition conditions"},
		{Name: "properties", Type: field.TypeJSON, Nullable: true, Comment: "Node properties"},
		{Name: "is_countersign", Type: field.TypeBool, Comment: "Whether requires countersign", Default: false},
		{Name: "countersign_rule", Type: field.TypeString, Nullable: true, Comment: "Countersign rules"},
		{Name: "handlers", Type: field.TypeJSON, Nullable: true, Comment: "Handler configurations"},
		{Name: "listeners", Type: field.TypeJSON, Nullable: true, Comment: "Listener configurations"},
		{Name: "hooks", Type: field.TypeJSON, Nullable: true, Comment: "Hook configurations"},
		{Name: "variables", Type: field.TypeJSON, Nullable: true, Comment: "Node variables"},
		{Name: "retry_times", Type: field.TypeInt, Nullable: true, Comment: "Number of retries", Default: 0},
		{Name: "retry_interval", Type: field.TypeInt, Nullable: true, Comment: "Retry interval in seconds", Default: 0},
		{Name: "is_working_day", Type: field.TypeBool, Comment: "Whether to count working days only", Default: true},
	}
	// NcseFlowNodeTable holds the schema information for the "ncse_flow_node" table.
	NcseFlowNodeTable = &schema.Table{
		Name:       "ncse_flow_node",
		Columns:    NcseFlowNodeColumns,
		PrimaryKey: []*schema.Column{NcseFlowNodeColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "node_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowNodeColumns[0]},
			},
			{
				Name:    "node_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowNodeColumns[36]},
			},
			{
				Name:    "node_node_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowNodeColumns[5]},
			},
			{
				Name:    "node_type",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowNodeColumns[3]},
			},
		},
	}
	// NcseFlowProcessColumns holds the columns for the "ncse_flow_process" table.
	NcseFlowProcessColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "process_id", Type: field.TypeString, Comment: "Process instance ID"},
		{Name: "template_id", Type: field.TypeString, Comment: "Process template ID"},
		{Name: "business_key", Type: field.TypeString, Comment: "Business document ID"},
		{Name: "form_code", Type: field.TypeString, Comment: "Form type code"},
		{Name: "form_version", Type: field.TypeString, Nullable: true, Comment: "Form version number"},
		{Name: "form_config", Type: field.TypeJSON, Nullable: true, Comment: "Form configuration"},
		{Name: "form_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Form permission settings"},
		{Name: "field_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Field level permissions"},
		{Name: "business_tags", Type: field.TypeJSON, Nullable: true, Comment: "Business tags"},
		{Name: "module_code", Type: field.TypeString, Comment: "Module code"},
		{Name: "category", Type: field.TypeString, Nullable: true, Comment: "Category"},
		{Name: "flow_status", Type: field.TypeString, Nullable: true, Comment: "Flow status"},
		{Name: "flow_variables", Type: field.TypeJSON, Nullable: true, Comment: "Flow variables"},
		{Name: "is_draft", Type: field.TypeBool, Comment: "Whether is draft", Default: false},
		{Name: "is_terminated", Type: field.TypeBool, Comment: "Whether is terminated", Default: false},
		{Name: "is_suspended", Type: field.TypeBool, Comment: "Whether is suspended", Default: false},
		{Name: "suspend_reason", Type: field.TypeString, Nullable: true, Comment: "Suspension reason"},
		{Name: "start_time", Type: field.TypeInt64, Comment: "Start time"},
		{Name: "end_time", Type: field.TypeInt64, Nullable: true, Comment: "End time"},
		{Name: "due_time", Type: field.TypeInt64, Nullable: true, Comment: "Due time"},
		{Name: "duration", Type: field.TypeInt, Nullable: true, Comment: "Duration in seconds"},
		{Name: "priority", Type: field.TypeInt, Comment: "Priority level", Default: 0},
		{Name: "is_timeout", Type: field.TypeBool, Comment: "Whether timed out", Default: false},
		{Name: "reminder_count", Type: field.TypeInt, Comment: "Number of reminders sent", Default: 0},
		{Name: "allow_cancel", Type: field.TypeBool, Comment: "Allow cancellation", Default: true},
		{Name: "allow_urge", Type: field.TypeBool, Comment: "Allow urging", Default: true},
		{Name: "allow_delegate", Type: field.TypeBool, Comment: "Allow delegation", Default: true},
		{Name: "allow_transfer", Type: field.TypeBool, Comment: "Allow transfer", Default: true},
		{Name: "is_draft_enabled", Type: field.TypeBool, Comment: "Whether draft is enabled", Default: true},
		{Name: "is_auto_start", Type: field.TypeBool, Comment: "Whether auto start is enabled", Default: false},
		{Name: "strict_mode", Type: field.TypeBool, Comment: "Enable strict mode", Default: false},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "process_key", Type: field.TypeString, Unique: true, Comment: "Process unique identifier"},
		{Name: "initiator", Type: field.TypeString, Comment: "Process initiator"},
		{Name: "initiator_dept", Type: field.TypeString, Nullable: true, Comment: "Initiator's department"},
		{Name: "process_code", Type: field.TypeString, Comment: "Process code"},
		{Name: "variables", Type: field.TypeJSON, Comment: "Process variables"},
		{Name: "current_node", Type: field.TypeString, Nullable: true, Comment: "Current node"},
		{Name: "active_nodes", Type: field.TypeJSON, Nullable: true, Comment: "Currently active nodes"},
		{Name: "process_snapshot", Type: field.TypeJSON, Nullable: true, Comment: "Process snapshot"},
		{Name: "form_snapshot", Type: field.TypeJSON, Nullable: true, Comment: "Form snapshot"},
		{Name: "urge_count", Type: field.TypeInt, Comment: "Number of urges", Default: 0},
	}
	// NcseFlowProcessTable holds the schema information for the "ncse_flow_process" table.
	NcseFlowProcessTable = &schema.Table{
		Name:       "ncse_flow_process",
		Columns:    NcseFlowProcessColumns,
		PrimaryKey: []*schema.Column{NcseFlowProcessColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "process_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowProcessColumns[0]},
			},
			{
				Name:    "process_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessColumns[34]},
			},
			{
				Name:    "process_process_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowProcessColumns[39]},
			},
			{
				Name:    "process_business_key",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessColumns[4]},
			},
			{
				Name:    "process_module_code_form_code",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessColumns[11], NcseFlowProcessColumns[5]},
			},
			{
				Name:    "process_initiator",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessColumns[40]},
			},
		},
	}
	// NcseFlowProcessDesignColumns holds the columns for the "ncse_flow_process_design" table.
	NcseFlowProcessDesignColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "version", Type: field.TypeString, Nullable: true, Comment: "Version"},
		{Name: "disabled", Type: field.TypeBool, Nullable: true, Comment: "is disabled", Default: false},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "template_id", Type: field.TypeString, Comment: "Template ID"},
		{Name: "graph_data", Type: field.TypeJSON, Nullable: true, Comment: "Process graph data"},
		{Name: "node_layouts", Type: field.TypeJSON, Nullable: true, Comment: "Node layout positions"},
		{Name: "properties", Type: field.TypeJSON, Nullable: true, Comment: "Process design properties"},
		{Name: "validation_rules", Type: field.TypeJSON, Nullable: true, Comment: "Process validation rules"},
		{Name: "is_draft", Type: field.TypeBool, Comment: "Whether is draft", Default: false},
		{Name: "source_version", Type: field.TypeString, Nullable: true, Comment: "Source version"},
	}
	// NcseFlowProcessDesignTable holds the schema information for the "ncse_flow_process_design" table.
	NcseFlowProcessDesignTable = &schema.Table{
		Name:       "ncse_flow_process_design",
		Columns:    NcseFlowProcessDesignColumns,
		PrimaryKey: []*schema.Column{NcseFlowProcessDesignColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "processdesign_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowProcessDesignColumns[0]},
			},
			{
				Name:    "processdesign_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessDesignColumns[4]},
			},
			{
				Name:    "processdesign_template_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowProcessDesignColumns[9]},
			},
		},
	}
	// NcseFlowRuleColumns holds the columns for the "ncse_flow_rule" table.
	NcseFlowRuleColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "code", Type: field.TypeString, Nullable: true, Comment: "code"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "rule_key", Type: field.TypeString, Unique: true, Comment: "Rule unique key"},
		{Name: "template_id", Type: field.TypeString, Nullable: true, Comment: "Template ID if template specific"},
		{Name: "node_key", Type: field.TypeString, Nullable: true, Comment: "Node key if node specific"},
		{Name: "conditions", Type: field.TypeJSON, Comment: "Rule conditions"},
		{Name: "actions", Type: field.TypeJSON, Comment: "Rule actions"},
		{Name: "priority", Type: field.TypeInt, Comment: "Rule priority", Default: 0},
		{Name: "is_enabled", Type: field.TypeBool, Comment: "Whether rule is enabled", Default: true},
		{Name: "effective_time", Type: field.TypeInt64, Nullable: true, Comment: "Rule effective time"},
		{Name: "expire_time", Type: field.TypeInt64, Nullable: true, Comment: "Rule expire time"},
	}
	// NcseFlowRuleTable holds the schema information for the "ncse_flow_rule" table.
	NcseFlowRuleTable = &schema.Table{
		Name:       "ncse_flow_rule",
		Columns:    NcseFlowRuleColumns,
		PrimaryKey: []*schema.Column{NcseFlowRuleColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "rule_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowRuleColumns[0]},
			},
			{
				Name:    "rule_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowRuleColumns[7]},
			},
			{
				Name:    "rule_rule_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowRuleColumns[12]},
			},
			{
				Name:    "rule_template_id_node_key",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowRuleColumns[13], NcseFlowRuleColumns[14]},
			},
		},
	}
	// NcseFlowTaskColumns holds the columns for the "ncse_flow_task" table.
	NcseFlowTaskColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "process_id", Type: field.TypeString, Comment: "Process instance ID"},
		{Name: "template_id", Type: field.TypeString, Comment: "Process template ID"},
		{Name: "business_key", Type: field.TypeString, Comment: "Business document ID"},
		{Name: "node_key", Type: field.TypeString, Unique: true, Comment: "Unique identifier for the node"},
		{Name: "node_type", Type: field.TypeString, Comment: "Node type"},
		{Name: "node_config", Type: field.TypeJSON, Nullable: true, Comment: "Node configuration"},
		{Name: "node_rules", Type: field.TypeJSON, Nullable: true, Comment: "Node rules"},
		{Name: "node_events", Type: field.TypeJSON, Nullable: true, Comment: "Node events"},
		{Name: "assignees", Type: field.TypeJSON, Comment: "Task assignees"},
		{Name: "candidates", Type: field.TypeJSON, Comment: "Candidate assignees"},
		{Name: "delegated_from", Type: field.TypeString, Nullable: true, Comment: "Delegated from user"},
		{Name: "delegated_reason", Type: field.TypeString, Nullable: true, Comment: "Delegation reason"},
		{Name: "is_delegated", Type: field.TypeBool, Comment: "Whether task is delegated", Default: false},
		{Name: "is_transferred", Type: field.TypeBool, Comment: "Whether task is transferred", Default: false},
		{Name: "start_time", Type: field.TypeInt64, Comment: "Start time"},
		{Name: "end_time", Type: field.TypeInt64, Nullable: true, Comment: "End time"},
		{Name: "due_time", Type: field.TypeInt64, Nullable: true, Comment: "Due time"},
		{Name: "duration", Type: field.TypeInt, Nullable: true, Comment: "Duration in seconds"},
		{Name: "priority", Type: field.TypeInt, Comment: "Priority level", Default: 0},
		{Name: "is_timeout", Type: field.TypeBool, Comment: "Whether timed out", Default: false},
		{Name: "reminder_count", Type: field.TypeInt, Comment: "Number of reminders sent", Default: 0},
		{Name: "allow_cancel", Type: field.TypeBool, Comment: "Allow cancellation", Default: true},
		{Name: "allow_urge", Type: field.TypeBool, Comment: "Allow urging", Default: true},
		{Name: "allow_delegate", Type: field.TypeBool, Comment: "Allow delegation", Default: true},
		{Name: "allow_transfer", Type: field.TypeBool, Comment: "Allow transfer", Default: true},
		{Name: "is_draft_enabled", Type: field.TypeBool, Comment: "Whether draft is enabled", Default: true},
		{Name: "is_auto_start", Type: field.TypeBool, Comment: "Whether auto start is enabled", Default: false},
		{Name: "strict_mode", Type: field.TypeBool, Comment: "Enable strict mode", Default: false},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "task_key", Type: field.TypeString, Unique: true, Comment: "Task unique identifier"},
		{Name: "parent_id", Type: field.TypeString, Nullable: true, Comment: "Parent task ID"},
		{Name: "child_ids", Type: field.TypeJSON, Comment: "Child task IDs"},
		{Name: "action", Type: field.TypeString, Nullable: true, Comment: "Processing action"},
		{Name: "comment", Type: field.TypeString, Nullable: true, Comment: "Processing comment"},
		{Name: "attachments", Type: field.TypeJSON, Nullable: true, Comment: "Attachment information"},
		{Name: "form_data", Type: field.TypeJSON, Nullable: true, Comment: "Form data"},
		{Name: "variables", Type: field.TypeJSON, Nullable: true, Comment: "Task variables"},
		{Name: "is_resubmit", Type: field.TypeBool, Comment: "Whether is resubmitted", Default: false},
		{Name: "claim_time", Type: field.TypeInt64, Nullable: true, Comment: "Claim time"},
		{Name: "is_urged", Type: field.TypeBool, Comment: "Whether is urged", Default: false},
		{Name: "urge_count", Type: field.TypeInt, Comment: "Number of urges", Default: 0},
	}
	// NcseFlowTaskTable holds the schema information for the "ncse_flow_task" table.
	NcseFlowTaskTable = &schema.Table{
		Name:       "ncse_flow_task",
		Columns:    NcseFlowTaskColumns,
		PrimaryKey: []*schema.Column{NcseFlowTaskColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "task_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowTaskColumns[0]},
			},
			{
				Name:    "task_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTaskColumns[33]},
			},
			{
				Name:    "task_task_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowTaskColumns[38]},
			},
			{
				Name:    "task_process_id_node_key",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTaskColumns[4], NcseFlowTaskColumns[7]},
			},
			{
				Name:    "task_node_type",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTaskColumns[8]},
			},
			{
				Name:    "task_due_time",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTaskColumns[20]},
			},
		},
	}
	// NcseFlowTemplateColumns holds the columns for the "ncse_flow_template" table.
	NcseFlowTemplateColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "code", Type: field.TypeString, Nullable: true, Comment: "code"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "version", Type: field.TypeString, Nullable: true, Comment: "Version"},
		{Name: "status", Type: field.TypeString, Nullable: true, Comment: "Status, text status"},
		{Name: "disabled", Type: field.TypeBool, Nullable: true, Comment: "is disabled", Default: false},
		{Name: "form_code", Type: field.TypeString, Comment: "Form type code"},
		{Name: "form_version", Type: field.TypeString, Nullable: true, Comment: "Form version number"},
		{Name: "form_config", Type: field.TypeJSON, Nullable: true, Comment: "Form configuration"},
		{Name: "form_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Form permission settings"},
		{Name: "field_permissions", Type: field.TypeJSON, Nullable: true, Comment: "Field level permissions"},
		{Name: "node_key", Type: field.TypeString, Unique: true, Comment: "Unique identifier for the node"},
		{Name: "node_type", Type: field.TypeString, Comment: "Node type"},
		{Name: "node_config", Type: field.TypeJSON, Nullable: true, Comment: "Node configuration"},
		{Name: "node_rules", Type: field.TypeJSON, Nullable: true, Comment: "Node rules"},
		{Name: "node_events", Type: field.TypeJSON, Nullable: true, Comment: "Node events"},
		{Name: "business_tags", Type: field.TypeJSON, Nullable: true, Comment: "Business tags"},
		{Name: "module_code", Type: field.TypeString, Comment: "Module code"},
		{Name: "category", Type: field.TypeString, Nullable: true, Comment: "Category"},
		{Name: "allow_cancel", Type: field.TypeBool, Comment: "Allow cancellation", Default: true},
		{Name: "allow_urge", Type: field.TypeBool, Comment: "Allow urging", Default: true},
		{Name: "allow_delegate", Type: field.TypeBool, Comment: "Allow delegation", Default: true},
		{Name: "allow_transfer", Type: field.TypeBool, Comment: "Allow transfer", Default: true},
		{Name: "is_draft_enabled", Type: field.TypeBool, Comment: "Whether draft is enabled", Default: true},
		{Name: "is_auto_start", Type: field.TypeBool, Comment: "Whether auto start is enabled", Default: false},
		{Name: "strict_mode", Type: field.TypeBool, Comment: "Enable strict mode", Default: false},
		{Name: "viewers", Type: field.TypeJSON, Nullable: true, Comment: "Users with view permission"},
		{Name: "editors", Type: field.TypeJSON, Nullable: true, Comment: "Users with edit permission"},
		{Name: "permission_configs", Type: field.TypeJSON, Nullable: true, Comment: "Permission configurations"},
		{Name: "role_configs", Type: field.TypeJSON, Nullable: true, Comment: "Role configurations"},
		{Name: "visible_range", Type: field.TypeJSON, Nullable: true, Comment: "Visibility range"},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
		{Name: "template_key", Type: field.TypeString, Unique: true, Comment: "Template unique identifier"},
		{Name: "process_rules", Type: field.TypeJSON, Nullable: true, Comment: "Process rules"},
		{Name: "trigger_conditions", Type: field.TypeJSON, Nullable: true, Comment: "Trigger conditions"},
		{Name: "timeout_config", Type: field.TypeJSON, Nullable: true, Comment: "Timeout configuration"},
		{Name: "reminder_config", Type: field.TypeJSON, Nullable: true, Comment: "Reminder configuration"},
		{Name: "source_version", Type: field.TypeString, Nullable: true, Comment: "Source version"},
		{Name: "is_latest", Type: field.TypeBool, Comment: "Whether is latest version", Default: false},
		{Name: "effective_time", Type: field.TypeInt64, Nullable: true, Comment: "Effective time"},
		{Name: "expire_time", Type: field.TypeInt64, Nullable: true, Comment: "Expire time"},
	}
	// NcseFlowTemplateTable holds the schema information for the "ncse_flow_template" table.
	NcseFlowTemplateTable = &schema.Table{
		Name:       "ncse_flow_template",
		Columns:    NcseFlowTemplateColumns,
		PrimaryKey: []*schema.Column{NcseFlowTemplateColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "template_id",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowTemplateColumns[0]},
			},
			{
				Name:    "template_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTemplateColumns[34]},
			},
			{
				Name:    "template_template_key",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowTemplateColumns[39]},
			},
			{
				Name:    "template_code",
				Unique:  true,
				Columns: []*schema.Column{NcseFlowTemplateColumns[2]},
			},
			{
				Name:    "template_module_code_form_code",
				Unique:  false,
				Columns: []*schema.Column{NcseFlowTemplateColumns[19], NcseFlowTemplateColumns[8]},
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		NcseFlowBusinessTable,
		NcseFlowDelegationTable,
		NcseFlowHistoryTable,
		NcseFlowNodeTable,
		NcseFlowProcessTable,
		NcseFlowProcessDesignTable,
		NcseFlowRuleTable,
		NcseFlowTaskTable,
		NcseFlowTemplateTable,
	}
)

func init() {
	NcseFlowBusinessTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_business",
	}
	NcseFlowDelegationTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_delegation",
	}
	NcseFlowHistoryTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_history",
	}
	NcseFlowNodeTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_node",
	}
	NcseFlowProcessTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_process",
	}
	NcseFlowProcessDesignTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_process_design",
	}
	NcseFlowRuleTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_rule",
	}
	NcseFlowTaskTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_task",
	}
	NcseFlowTemplateTable.Annotation = &entsql.Annotation{
		Table: "ncse_flow_template",
	}
}
