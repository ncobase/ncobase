// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"ncobase/tenant/data/ent/tenantbilling"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// TenantBillingCreate is the builder for creating a TenantBilling entity.
type TenantBillingCreate struct {
	config
	mutation *TenantBillingMutation
	hooks    []Hook
}

// SetTenantID sets the "tenant_id" field.
func (tbc *TenantBillingCreate) SetTenantID(s string) *TenantBillingCreate {
	tbc.mutation.SetTenantID(s)
	return tbc
}

// SetNillableTenantID sets the "tenant_id" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableTenantID(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetTenantID(*s)
	}
	return tbc
}

// SetDescription sets the "description" field.
func (tbc *TenantBillingCreate) SetDescription(s string) *TenantBillingCreate {
	tbc.mutation.SetDescription(s)
	return tbc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableDescription(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetDescription(*s)
	}
	return tbc
}

// SetExtras sets the "extras" field.
func (tbc *TenantBillingCreate) SetExtras(m map[string]interface{}) *TenantBillingCreate {
	tbc.mutation.SetExtras(m)
	return tbc
}

// SetCreatedBy sets the "created_by" field.
func (tbc *TenantBillingCreate) SetCreatedBy(s string) *TenantBillingCreate {
	tbc.mutation.SetCreatedBy(s)
	return tbc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableCreatedBy(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetCreatedBy(*s)
	}
	return tbc
}

// SetUpdatedBy sets the "updated_by" field.
func (tbc *TenantBillingCreate) SetUpdatedBy(s string) *TenantBillingCreate {
	tbc.mutation.SetUpdatedBy(s)
	return tbc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableUpdatedBy(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetUpdatedBy(*s)
	}
	return tbc
}

// SetCreatedAt sets the "created_at" field.
func (tbc *TenantBillingCreate) SetCreatedAt(i int64) *TenantBillingCreate {
	tbc.mutation.SetCreatedAt(i)
	return tbc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableCreatedAt(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetCreatedAt(*i)
	}
	return tbc
}

// SetUpdatedAt sets the "updated_at" field.
func (tbc *TenantBillingCreate) SetUpdatedAt(i int64) *TenantBillingCreate {
	tbc.mutation.SetUpdatedAt(i)
	return tbc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableUpdatedAt(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetUpdatedAt(*i)
	}
	return tbc
}

// SetBillingPeriod sets the "billing_period" field.
func (tbc *TenantBillingCreate) SetBillingPeriod(s string) *TenantBillingCreate {
	tbc.mutation.SetBillingPeriod(s)
	return tbc
}

// SetNillableBillingPeriod sets the "billing_period" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableBillingPeriod(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetBillingPeriod(*s)
	}
	return tbc
}

// SetPeriodStart sets the "period_start" field.
func (tbc *TenantBillingCreate) SetPeriodStart(i int64) *TenantBillingCreate {
	tbc.mutation.SetPeriodStart(i)
	return tbc
}

// SetNillablePeriodStart sets the "period_start" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillablePeriodStart(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetPeriodStart(*i)
	}
	return tbc
}

// SetPeriodEnd sets the "period_end" field.
func (tbc *TenantBillingCreate) SetPeriodEnd(i int64) *TenantBillingCreate {
	tbc.mutation.SetPeriodEnd(i)
	return tbc
}

// SetNillablePeriodEnd sets the "period_end" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillablePeriodEnd(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetPeriodEnd(*i)
	}
	return tbc
}

// SetAmount sets the "amount" field.
func (tbc *TenantBillingCreate) SetAmount(f float64) *TenantBillingCreate {
	tbc.mutation.SetAmount(f)
	return tbc
}

// SetCurrency sets the "currency" field.
func (tbc *TenantBillingCreate) SetCurrency(s string) *TenantBillingCreate {
	tbc.mutation.SetCurrency(s)
	return tbc
}

// SetNillableCurrency sets the "currency" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableCurrency(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetCurrency(*s)
	}
	return tbc
}

// SetStatus sets the "status" field.
func (tbc *TenantBillingCreate) SetStatus(s string) *TenantBillingCreate {
	tbc.mutation.SetStatus(s)
	return tbc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableStatus(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetStatus(*s)
	}
	return tbc
}

// SetInvoiceNumber sets the "invoice_number" field.
func (tbc *TenantBillingCreate) SetInvoiceNumber(s string) *TenantBillingCreate {
	tbc.mutation.SetInvoiceNumber(s)
	return tbc
}

// SetNillableInvoiceNumber sets the "invoice_number" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableInvoiceNumber(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetInvoiceNumber(*s)
	}
	return tbc
}

// SetPaymentMethod sets the "payment_method" field.
func (tbc *TenantBillingCreate) SetPaymentMethod(s string) *TenantBillingCreate {
	tbc.mutation.SetPaymentMethod(s)
	return tbc
}

// SetNillablePaymentMethod sets the "payment_method" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillablePaymentMethod(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetPaymentMethod(*s)
	}
	return tbc
}

// SetPaidAt sets the "paid_at" field.
func (tbc *TenantBillingCreate) SetPaidAt(i int64) *TenantBillingCreate {
	tbc.mutation.SetPaidAt(i)
	return tbc
}

// SetNillablePaidAt sets the "paid_at" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillablePaidAt(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetPaidAt(*i)
	}
	return tbc
}

// SetDueDate sets the "due_date" field.
func (tbc *TenantBillingCreate) SetDueDate(i int64) *TenantBillingCreate {
	tbc.mutation.SetDueDate(i)
	return tbc
}

// SetNillableDueDate sets the "due_date" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableDueDate(i *int64) *TenantBillingCreate {
	if i != nil {
		tbc.SetDueDate(*i)
	}
	return tbc
}

// SetUsageDetails sets the "usage_details" field.
func (tbc *TenantBillingCreate) SetUsageDetails(m map[string]interface{}) *TenantBillingCreate {
	tbc.mutation.SetUsageDetails(m)
	return tbc
}

// SetID sets the "id" field.
func (tbc *TenantBillingCreate) SetID(s string) *TenantBillingCreate {
	tbc.mutation.SetID(s)
	return tbc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tbc *TenantBillingCreate) SetNillableID(s *string) *TenantBillingCreate {
	if s != nil {
		tbc.SetID(*s)
	}
	return tbc
}

// Mutation returns the TenantBillingMutation object of the builder.
func (tbc *TenantBillingCreate) Mutation() *TenantBillingMutation {
	return tbc.mutation
}

// Save creates the TenantBilling in the database.
func (tbc *TenantBillingCreate) Save(ctx context.Context) (*TenantBilling, error) {
	tbc.defaults()
	return withHooks(ctx, tbc.sqlSave, tbc.mutation, tbc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tbc *TenantBillingCreate) SaveX(ctx context.Context) *TenantBilling {
	v, err := tbc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tbc *TenantBillingCreate) Exec(ctx context.Context) error {
	_, err := tbc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tbc *TenantBillingCreate) ExecX(ctx context.Context) {
	if err := tbc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tbc *TenantBillingCreate) defaults() {
	if _, ok := tbc.mutation.Extras(); !ok {
		v := tenantbilling.DefaultExtras
		tbc.mutation.SetExtras(v)
	}
	if _, ok := tbc.mutation.CreatedAt(); !ok {
		v := tenantbilling.DefaultCreatedAt()
		tbc.mutation.SetCreatedAt(v)
	}
	if _, ok := tbc.mutation.UpdatedAt(); !ok {
		v := tenantbilling.DefaultUpdatedAt()
		tbc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tbc.mutation.BillingPeriod(); !ok {
		v := tenantbilling.DefaultBillingPeriod
		tbc.mutation.SetBillingPeriod(v)
	}
	if _, ok := tbc.mutation.Currency(); !ok {
		v := tenantbilling.DefaultCurrency
		tbc.mutation.SetCurrency(v)
	}
	if _, ok := tbc.mutation.Status(); !ok {
		v := tenantbilling.DefaultStatus
		tbc.mutation.SetStatus(v)
	}
	if _, ok := tbc.mutation.ID(); !ok {
		v := tenantbilling.DefaultID()
		tbc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tbc *TenantBillingCreate) check() error {
	if _, ok := tbc.mutation.BillingPeriod(); !ok {
		return &ValidationError{Name: "billing_period", err: errors.New(`ent: missing required field "TenantBilling.billing_period"`)}
	}
	if _, ok := tbc.mutation.Amount(); !ok {
		return &ValidationError{Name: "amount", err: errors.New(`ent: missing required field "TenantBilling.amount"`)}
	}
	if v, ok := tbc.mutation.Amount(); ok {
		if err := tenantbilling.AmountValidator(v); err != nil {
			return &ValidationError{Name: "amount", err: fmt.Errorf(`ent: validator failed for field "TenantBilling.amount": %w`, err)}
		}
	}
	if _, ok := tbc.mutation.Currency(); !ok {
		return &ValidationError{Name: "currency", err: errors.New(`ent: missing required field "TenantBilling.currency"`)}
	}
	if _, ok := tbc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "TenantBilling.status"`)}
	}
	if v, ok := tbc.mutation.ID(); ok {
		if err := tenantbilling.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "TenantBilling.id": %w`, err)}
		}
	}
	return nil
}

func (tbc *TenantBillingCreate) sqlSave(ctx context.Context) (*TenantBilling, error) {
	if err := tbc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tbc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tbc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected TenantBilling.ID type: %T", _spec.ID.Value)
		}
	}
	tbc.mutation.id = &_node.ID
	tbc.mutation.done = true
	return _node, nil
}

func (tbc *TenantBillingCreate) createSpec() (*TenantBilling, *sqlgraph.CreateSpec) {
	var (
		_node = &TenantBilling{config: tbc.config}
		_spec = sqlgraph.NewCreateSpec(tenantbilling.Table, sqlgraph.NewFieldSpec(tenantbilling.FieldID, field.TypeString))
	)
	if id, ok := tbc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tbc.mutation.TenantID(); ok {
		_spec.SetField(tenantbilling.FieldTenantID, field.TypeString, value)
		_node.TenantID = value
	}
	if value, ok := tbc.mutation.Description(); ok {
		_spec.SetField(tenantbilling.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := tbc.mutation.Extras(); ok {
		_spec.SetField(tenantbilling.FieldExtras, field.TypeJSON, value)
		_node.Extras = value
	}
	if value, ok := tbc.mutation.CreatedBy(); ok {
		_spec.SetField(tenantbilling.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := tbc.mutation.UpdatedBy(); ok {
		_spec.SetField(tenantbilling.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := tbc.mutation.CreatedAt(); ok {
		_spec.SetField(tenantbilling.FieldCreatedAt, field.TypeInt64, value)
		_node.CreatedAt = value
	}
	if value, ok := tbc.mutation.UpdatedAt(); ok {
		_spec.SetField(tenantbilling.FieldUpdatedAt, field.TypeInt64, value)
		_node.UpdatedAt = value
	}
	if value, ok := tbc.mutation.BillingPeriod(); ok {
		_spec.SetField(tenantbilling.FieldBillingPeriod, field.TypeString, value)
		_node.BillingPeriod = value
	}
	if value, ok := tbc.mutation.PeriodStart(); ok {
		_spec.SetField(tenantbilling.FieldPeriodStart, field.TypeInt64, value)
		_node.PeriodStart = value
	}
	if value, ok := tbc.mutation.PeriodEnd(); ok {
		_spec.SetField(tenantbilling.FieldPeriodEnd, field.TypeInt64, value)
		_node.PeriodEnd = value
	}
	if value, ok := tbc.mutation.Amount(); ok {
		_spec.SetField(tenantbilling.FieldAmount, field.TypeFloat64, value)
		_node.Amount = value
	}
	if value, ok := tbc.mutation.Currency(); ok {
		_spec.SetField(tenantbilling.FieldCurrency, field.TypeString, value)
		_node.Currency = value
	}
	if value, ok := tbc.mutation.Status(); ok {
		_spec.SetField(tenantbilling.FieldStatus, field.TypeString, value)
		_node.Status = value
	}
	if value, ok := tbc.mutation.InvoiceNumber(); ok {
		_spec.SetField(tenantbilling.FieldInvoiceNumber, field.TypeString, value)
		_node.InvoiceNumber = value
	}
	if value, ok := tbc.mutation.PaymentMethod(); ok {
		_spec.SetField(tenantbilling.FieldPaymentMethod, field.TypeString, value)
		_node.PaymentMethod = value
	}
	if value, ok := tbc.mutation.PaidAt(); ok {
		_spec.SetField(tenantbilling.FieldPaidAt, field.TypeInt64, value)
		_node.PaidAt = value
	}
	if value, ok := tbc.mutation.DueDate(); ok {
		_spec.SetField(tenantbilling.FieldDueDate, field.TypeInt64, value)
		_node.DueDate = value
	}
	if value, ok := tbc.mutation.UsageDetails(); ok {
		_spec.SetField(tenantbilling.FieldUsageDetails, field.TypeJSON, value)
		_node.UsageDetails = value
	}
	return _node, _spec
}

// TenantBillingCreateBulk is the builder for creating many TenantBilling entities in bulk.
type TenantBillingCreateBulk struct {
	config
	err      error
	builders []*TenantBillingCreate
}

// Save creates the TenantBilling entities in the database.
func (tbcb *TenantBillingCreateBulk) Save(ctx context.Context) ([]*TenantBilling, error) {
	if tbcb.err != nil {
		return nil, tbcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tbcb.builders))
	nodes := make([]*TenantBilling, len(tbcb.builders))
	mutators := make([]Mutator, len(tbcb.builders))
	for i := range tbcb.builders {
		func(i int, root context.Context) {
			builder := tbcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TenantBillingMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, tbcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tbcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, tbcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tbcb *TenantBillingCreateBulk) SaveX(ctx context.Context) []*TenantBilling {
	v, err := tbcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tbcb *TenantBillingCreateBulk) Exec(ctx context.Context) error {
	_, err := tbcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tbcb *TenantBillingCreateBulk) ExecX(ctx context.Context) {
	if err := tbcb.Exec(ctx); err != nil {
		panic(err)
	}
}
