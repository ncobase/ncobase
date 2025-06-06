// Code generated by ent, DO NOT EDIT.

package ent

import (
	"ncobase/payment/data/ent/paymentchannel"
	"ncobase/payment/data/ent/paymentlog"
	"ncobase/payment/data/ent/paymentorder"
	"ncobase/payment/data/ent/paymentproduct"
	"ncobase/payment/data/ent/paymentsubscription"
	"ncobase/payment/data/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	paymentchannelMixin := schema.PaymentChannel{}.Mixin()
	paymentchannelMixinFields0 := paymentchannelMixin[0].Fields()
	_ = paymentchannelMixinFields0
	paymentchannelMixinFields3 := paymentchannelMixin[3].Fields()
	_ = paymentchannelMixinFields3
	paymentchannelMixinFields5 := paymentchannelMixin[5].Fields()
	_ = paymentchannelMixinFields5
	paymentchannelFields := schema.PaymentChannel{}.Fields()
	_ = paymentchannelFields
	// paymentchannelDescExtras is the schema descriptor for extras field.
	paymentchannelDescExtras := paymentchannelMixinFields3[0].Descriptor()
	// paymentchannel.DefaultExtras holds the default value on creation for the extras field.
	paymentchannel.DefaultExtras = paymentchannelDescExtras.Default.(map[string]interface{})
	// paymentchannelDescCreatedAt is the schema descriptor for created_at field.
	paymentchannelDescCreatedAt := paymentchannelMixinFields5[0].Descriptor()
	// paymentchannel.DefaultCreatedAt holds the default value on creation for the created_at field.
	paymentchannel.DefaultCreatedAt = paymentchannelDescCreatedAt.Default.(func() int64)
	// paymentchannelDescUpdatedAt is the schema descriptor for updated_at field.
	paymentchannelDescUpdatedAt := paymentchannelMixinFields5[1].Descriptor()
	// paymentchannel.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	paymentchannel.DefaultUpdatedAt = paymentchannelDescUpdatedAt.Default.(func() int64)
	// paymentchannel.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	paymentchannel.UpdateDefaultUpdatedAt = paymentchannelDescUpdatedAt.UpdateDefault.(func() int64)
	// paymentchannelDescProvider is the schema descriptor for provider field.
	paymentchannelDescProvider := paymentchannelFields[0].Descriptor()
	// paymentchannel.ProviderValidator is a validator for the "provider" field. It is called by the builders before save.
	paymentchannel.ProviderValidator = paymentchannelDescProvider.Validators[0].(func(string) error)
	// paymentchannelDescStatus is the schema descriptor for status field.
	paymentchannelDescStatus := paymentchannelFields[1].Descriptor()
	// paymentchannel.DefaultStatus holds the default value on creation for the status field.
	paymentchannel.DefaultStatus = paymentchannelDescStatus.Default.(string)
	// paymentchannelDescIsDefault is the schema descriptor for is_default field.
	paymentchannelDescIsDefault := paymentchannelFields[2].Descriptor()
	// paymentchannel.DefaultIsDefault holds the default value on creation for the is_default field.
	paymentchannel.DefaultIsDefault = paymentchannelDescIsDefault.Default.(bool)
	// paymentchannelDescID is the schema descriptor for id field.
	paymentchannelDescID := paymentchannelMixinFields0[0].Descriptor()
	// paymentchannel.DefaultID holds the default value on creation for the id field.
	paymentchannel.DefaultID = paymentchannelDescID.Default.(func() string)
	// paymentchannel.IDValidator is a validator for the "id" field. It is called by the builders before save.
	paymentchannel.IDValidator = paymentchannelDescID.Validators[0].(func(string) error)
	paymentlogMixin := schema.PaymentLog{}.Mixin()
	paymentlogMixinFields0 := paymentlogMixin[0].Fields()
	_ = paymentlogMixinFields0
	paymentlogMixinFields1 := paymentlogMixin[1].Fields()
	_ = paymentlogMixinFields1
	paymentlogMixinFields2 := paymentlogMixin[2].Fields()
	_ = paymentlogMixinFields2
	paymentlogFields := schema.PaymentLog{}.Fields()
	_ = paymentlogFields
	// paymentlogDescExtras is the schema descriptor for extras field.
	paymentlogDescExtras := paymentlogMixinFields1[0].Descriptor()
	// paymentlog.DefaultExtras holds the default value on creation for the extras field.
	paymentlog.DefaultExtras = paymentlogDescExtras.Default.(map[string]interface{})
	// paymentlogDescCreatedAt is the schema descriptor for created_at field.
	paymentlogDescCreatedAt := paymentlogMixinFields2[0].Descriptor()
	// paymentlog.DefaultCreatedAt holds the default value on creation for the created_at field.
	paymentlog.DefaultCreatedAt = paymentlogDescCreatedAt.Default.(func() int64)
	// paymentlogDescUpdatedAt is the schema descriptor for updated_at field.
	paymentlogDescUpdatedAt := paymentlogMixinFields2[1].Descriptor()
	// paymentlog.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	paymentlog.DefaultUpdatedAt = paymentlogDescUpdatedAt.Default.(func() int64)
	// paymentlog.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	paymentlog.UpdateDefaultUpdatedAt = paymentlogDescUpdatedAt.UpdateDefault.(func() int64)
	// paymentlogDescOrderID is the schema descriptor for order_id field.
	paymentlogDescOrderID := paymentlogFields[0].Descriptor()
	// paymentlog.OrderIDValidator is a validator for the "order_id" field. It is called by the builders before save.
	paymentlog.OrderIDValidator = paymentlogDescOrderID.Validators[0].(func(string) error)
	// paymentlogDescChannelID is the schema descriptor for channel_id field.
	paymentlogDescChannelID := paymentlogFields[1].Descriptor()
	// paymentlog.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	paymentlog.ChannelIDValidator = paymentlogDescChannelID.Validators[0].(func(string) error)
	// paymentlogDescType is the schema descriptor for type field.
	paymentlogDescType := paymentlogFields[2].Descriptor()
	// paymentlog.TypeValidator is a validator for the "type" field. It is called by the builders before save.
	paymentlog.TypeValidator = paymentlogDescType.Validators[0].(func(string) error)
	// paymentlogDescID is the schema descriptor for id field.
	paymentlogDescID := paymentlogMixinFields0[0].Descriptor()
	// paymentlog.DefaultID holds the default value on creation for the id field.
	paymentlog.DefaultID = paymentlogDescID.Default.(func() string)
	// paymentlog.IDValidator is a validator for the "id" field. It is called by the builders before save.
	paymentlog.IDValidator = paymentlogDescID.Validators[0].(func(string) error)
	paymentorderMixin := schema.PaymentOrder{}.Mixin()
	paymentorderMixinFields0 := paymentorderMixin[0].Fields()
	_ = paymentorderMixinFields0
	paymentorderMixinFields1 := paymentorderMixin[1].Fields()
	_ = paymentorderMixinFields1
	paymentorderMixinFields3 := paymentorderMixin[3].Fields()
	_ = paymentorderMixinFields3
	paymentorderFields := schema.PaymentOrder{}.Fields()
	_ = paymentorderFields
	// paymentorderDescExtras is the schema descriptor for extras field.
	paymentorderDescExtras := paymentorderMixinFields1[0].Descriptor()
	// paymentorder.DefaultExtras holds the default value on creation for the extras field.
	paymentorder.DefaultExtras = paymentorderDescExtras.Default.(map[string]interface{})
	// paymentorderDescCreatedAt is the schema descriptor for created_at field.
	paymentorderDescCreatedAt := paymentorderMixinFields3[0].Descriptor()
	// paymentorder.DefaultCreatedAt holds the default value on creation for the created_at field.
	paymentorder.DefaultCreatedAt = paymentorderDescCreatedAt.Default.(func() int64)
	// paymentorderDescUpdatedAt is the schema descriptor for updated_at field.
	paymentorderDescUpdatedAt := paymentorderMixinFields3[1].Descriptor()
	// paymentorder.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	paymentorder.DefaultUpdatedAt = paymentorderDescUpdatedAt.Default.(func() int64)
	// paymentorder.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	paymentorder.UpdateDefaultUpdatedAt = paymentorderDescUpdatedAt.UpdateDefault.(func() int64)
	// paymentorderDescOrderNumber is the schema descriptor for order_number field.
	paymentorderDescOrderNumber := paymentorderFields[0].Descriptor()
	// paymentorder.OrderNumberValidator is a validator for the "order_number" field. It is called by the builders before save.
	paymentorder.OrderNumberValidator = paymentorderDescOrderNumber.Validators[0].(func(string) error)
	// paymentorderDescAmount is the schema descriptor for amount field.
	paymentorderDescAmount := paymentorderFields[1].Descriptor()
	// paymentorder.AmountValidator is a validator for the "amount" field. It is called by the builders before save.
	paymentorder.AmountValidator = paymentorderDescAmount.Validators[0].(func(float64) error)
	// paymentorderDescCurrency is the schema descriptor for currency field.
	paymentorderDescCurrency := paymentorderFields[2].Descriptor()
	// paymentorder.DefaultCurrency holds the default value on creation for the currency field.
	paymentorder.DefaultCurrency = paymentorderDescCurrency.Default.(string)
	// paymentorderDescStatus is the schema descriptor for status field.
	paymentorderDescStatus := paymentorderFields[3].Descriptor()
	// paymentorder.DefaultStatus holds the default value on creation for the status field.
	paymentorder.DefaultStatus = paymentorderDescStatus.Default.(string)
	// paymentorderDescType is the schema descriptor for type field.
	paymentorderDescType := paymentorderFields[4].Descriptor()
	// paymentorder.DefaultType holds the default value on creation for the type field.
	paymentorder.DefaultType = paymentorderDescType.Default.(string)
	// paymentorderDescChannelID is the schema descriptor for channel_id field.
	paymentorderDescChannelID := paymentorderFields[5].Descriptor()
	// paymentorder.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	paymentorder.ChannelIDValidator = paymentorderDescChannelID.Validators[0].(func(string) error)
	// paymentorderDescUserID is the schema descriptor for user_id field.
	paymentorderDescUserID := paymentorderFields[6].Descriptor()
	// paymentorder.UserIDValidator is a validator for the "user_id" field. It is called by the builders before save.
	paymentorder.UserIDValidator = paymentorderDescUserID.Validators[0].(func(string) error)
	// paymentorderDescID is the schema descriptor for id field.
	paymentorderDescID := paymentorderMixinFields0[0].Descriptor()
	// paymentorder.DefaultID holds the default value on creation for the id field.
	paymentorder.DefaultID = paymentorderDescID.Default.(func() string)
	// paymentorder.IDValidator is a validator for the "id" field. It is called by the builders before save.
	paymentorder.IDValidator = paymentorderDescID.Validators[0].(func(string) error)
	paymentproductMixin := schema.PaymentProduct{}.Mixin()
	paymentproductMixinFields0 := paymentproductMixin[0].Fields()
	_ = paymentproductMixinFields0
	paymentproductMixinFields3 := paymentproductMixin[3].Fields()
	_ = paymentproductMixinFields3
	paymentproductMixinFields5 := paymentproductMixin[5].Fields()
	_ = paymentproductMixinFields5
	paymentproductFields := schema.PaymentProduct{}.Fields()
	_ = paymentproductFields
	// paymentproductDescExtras is the schema descriptor for extras field.
	paymentproductDescExtras := paymentproductMixinFields3[0].Descriptor()
	// paymentproduct.DefaultExtras holds the default value on creation for the extras field.
	paymentproduct.DefaultExtras = paymentproductDescExtras.Default.(map[string]interface{})
	// paymentproductDescCreatedAt is the schema descriptor for created_at field.
	paymentproductDescCreatedAt := paymentproductMixinFields5[0].Descriptor()
	// paymentproduct.DefaultCreatedAt holds the default value on creation for the created_at field.
	paymentproduct.DefaultCreatedAt = paymentproductDescCreatedAt.Default.(func() int64)
	// paymentproductDescUpdatedAt is the schema descriptor for updated_at field.
	paymentproductDescUpdatedAt := paymentproductMixinFields5[1].Descriptor()
	// paymentproduct.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	paymentproduct.DefaultUpdatedAt = paymentproductDescUpdatedAt.Default.(func() int64)
	// paymentproduct.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	paymentproduct.UpdateDefaultUpdatedAt = paymentproductDescUpdatedAt.UpdateDefault.(func() int64)
	// paymentproductDescStatus is the schema descriptor for status field.
	paymentproductDescStatus := paymentproductFields[0].Descriptor()
	// paymentproduct.DefaultStatus holds the default value on creation for the status field.
	paymentproduct.DefaultStatus = paymentproductDescStatus.Default.(string)
	// paymentproductDescPricingType is the schema descriptor for pricing_type field.
	paymentproductDescPricingType := paymentproductFields[1].Descriptor()
	// paymentproduct.DefaultPricingType holds the default value on creation for the pricing_type field.
	paymentproduct.DefaultPricingType = paymentproductDescPricingType.Default.(string)
	// paymentproductDescPrice is the schema descriptor for price field.
	paymentproductDescPrice := paymentproductFields[2].Descriptor()
	// paymentproduct.PriceValidator is a validator for the "price" field. It is called by the builders before save.
	paymentproduct.PriceValidator = paymentproductDescPrice.Validators[0].(func(float64) error)
	// paymentproductDescCurrency is the schema descriptor for currency field.
	paymentproductDescCurrency := paymentproductFields[3].Descriptor()
	// paymentproduct.DefaultCurrency holds the default value on creation for the currency field.
	paymentproduct.DefaultCurrency = paymentproductDescCurrency.Default.(string)
	// paymentproductDescTrialDays is the schema descriptor for trial_days field.
	paymentproductDescTrialDays := paymentproductFields[5].Descriptor()
	// paymentproduct.DefaultTrialDays holds the default value on creation for the trial_days field.
	paymentproduct.DefaultTrialDays = paymentproductDescTrialDays.Default.(int)
	// paymentproductDescID is the schema descriptor for id field.
	paymentproductDescID := paymentproductMixinFields0[0].Descriptor()
	// paymentproduct.DefaultID holds the default value on creation for the id field.
	paymentproduct.DefaultID = paymentproductDescID.Default.(func() string)
	// paymentproduct.IDValidator is a validator for the "id" field. It is called by the builders before save.
	paymentproduct.IDValidator = paymentproductDescID.Validators[0].(func(string) error)
	paymentsubscriptionMixin := schema.PaymentSubscription{}.Mixin()
	paymentsubscriptionMixinFields0 := paymentsubscriptionMixin[0].Fields()
	_ = paymentsubscriptionMixinFields0
	paymentsubscriptionMixinFields1 := paymentsubscriptionMixin[1].Fields()
	_ = paymentsubscriptionMixinFields1
	paymentsubscriptionMixinFields3 := paymentsubscriptionMixin[3].Fields()
	_ = paymentsubscriptionMixinFields3
	paymentsubscriptionFields := schema.PaymentSubscription{}.Fields()
	_ = paymentsubscriptionFields
	// paymentsubscriptionDescExtras is the schema descriptor for extras field.
	paymentsubscriptionDescExtras := paymentsubscriptionMixinFields1[0].Descriptor()
	// paymentsubscription.DefaultExtras holds the default value on creation for the extras field.
	paymentsubscription.DefaultExtras = paymentsubscriptionDescExtras.Default.(map[string]interface{})
	// paymentsubscriptionDescCreatedAt is the schema descriptor for created_at field.
	paymentsubscriptionDescCreatedAt := paymentsubscriptionMixinFields3[0].Descriptor()
	// paymentsubscription.DefaultCreatedAt holds the default value on creation for the created_at field.
	paymentsubscription.DefaultCreatedAt = paymentsubscriptionDescCreatedAt.Default.(func() int64)
	// paymentsubscriptionDescUpdatedAt is the schema descriptor for updated_at field.
	paymentsubscriptionDescUpdatedAt := paymentsubscriptionMixinFields3[1].Descriptor()
	// paymentsubscription.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	paymentsubscription.DefaultUpdatedAt = paymentsubscriptionDescUpdatedAt.Default.(func() int64)
	// paymentsubscription.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	paymentsubscription.UpdateDefaultUpdatedAt = paymentsubscriptionDescUpdatedAt.UpdateDefault.(func() int64)
	// paymentsubscriptionDescStatus is the schema descriptor for status field.
	paymentsubscriptionDescStatus := paymentsubscriptionFields[0].Descriptor()
	// paymentsubscription.DefaultStatus holds the default value on creation for the status field.
	paymentsubscription.DefaultStatus = paymentsubscriptionDescStatus.Default.(string)
	// paymentsubscriptionDescUserID is the schema descriptor for user_id field.
	paymentsubscriptionDescUserID := paymentsubscriptionFields[1].Descriptor()
	// paymentsubscription.UserIDValidator is a validator for the "user_id" field. It is called by the builders before save.
	paymentsubscription.UserIDValidator = paymentsubscriptionDescUserID.Validators[0].(func(string) error)
	// paymentsubscriptionDescProductID is the schema descriptor for product_id field.
	paymentsubscriptionDescProductID := paymentsubscriptionFields[3].Descriptor()
	// paymentsubscription.ProductIDValidator is a validator for the "product_id" field. It is called by the builders before save.
	paymentsubscription.ProductIDValidator = paymentsubscriptionDescProductID.Validators[0].(func(string) error)
	// paymentsubscriptionDescChannelID is the schema descriptor for channel_id field.
	paymentsubscriptionDescChannelID := paymentsubscriptionFields[4].Descriptor()
	// paymentsubscription.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	paymentsubscription.ChannelIDValidator = paymentsubscriptionDescChannelID.Validators[0].(func(string) error)
	// paymentsubscriptionDescID is the schema descriptor for id field.
	paymentsubscriptionDescID := paymentsubscriptionMixinFields0[0].Descriptor()
	// paymentsubscription.DefaultID holds the default value on creation for the id field.
	paymentsubscription.DefaultID = paymentsubscriptionDescID.Default.(func() string)
	// paymentsubscription.IDValidator is a validator for the "id" field. It is called by the builders before save.
	paymentsubscription.IDValidator = paymentsubscriptionDescID.Validators[0].(func(string) error)
}
