// Code generated by ent, DO NOT EDIT.

package ent

import (
	"ncobase/core/realtime/data/ent/channel"
	"ncobase/core/realtime/data/ent/event"
	"ncobase/core/realtime/data/ent/notification"
	"ncobase/core/realtime/data/ent/subscription"
	"ncobase/core/realtime/data/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	channelMixin := schema.Channel{}.Mixin()
	channelMixinFields0 := channelMixin[0].Fields()
	_ = channelMixinFields0
	channelMixinFields4 := channelMixin[4].Fields()
	_ = channelMixinFields4
	channelMixinFields5 := channelMixin[5].Fields()
	_ = channelMixinFields5
	channelMixinFields6 := channelMixin[6].Fields()
	_ = channelMixinFields6
	channelFields := schema.Channel{}.Fields()
	_ = channelFields
	// channelDescStatus is the schema descriptor for status field.
	channelDescStatus := channelMixinFields4[0].Descriptor()
	// channel.DefaultStatus holds the default value on creation for the status field.
	channel.DefaultStatus = channelDescStatus.Default.(int)
	// channelDescExtras is the schema descriptor for extras field.
	channelDescExtras := channelMixinFields5[0].Descriptor()
	// channel.DefaultExtras holds the default value on creation for the extras field.
	channel.DefaultExtras = channelDescExtras.Default.(map[string]interface{})
	// channelDescCreatedAt is the schema descriptor for created_at field.
	channelDescCreatedAt := channelMixinFields6[0].Descriptor()
	// channel.DefaultCreatedAt holds the default value on creation for the created_at field.
	channel.DefaultCreatedAt = channelDescCreatedAt.Default.(func() int64)
	// channelDescUpdatedAt is the schema descriptor for updated_at field.
	channelDescUpdatedAt := channelMixinFields6[1].Descriptor()
	// channel.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	channel.DefaultUpdatedAt = channelDescUpdatedAt.Default.(func() int64)
	// channel.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	channel.UpdateDefaultUpdatedAt = channelDescUpdatedAt.UpdateDefault.(func() int64)
	// channelDescID is the schema descriptor for id field.
	channelDescID := channelMixinFields0[0].Descriptor()
	// channel.DefaultID holds the default value on creation for the id field.
	channel.DefaultID = channelDescID.Default.(func() string)
	// channel.IDValidator is a validator for the "id" field. It is called by the builders before save.
	channel.IDValidator = channelDescID.Validators[0].(func(string) error)
	eventMixin := schema.Event{}.Mixin()
	eventMixinFields0 := eventMixin[0].Fields()
	_ = eventMixinFields0
	eventMixinFields2 := eventMixin[2].Fields()
	_ = eventMixinFields2
	eventMixinFields3 := eventMixin[3].Fields()
	_ = eventMixinFields3
	eventMixinFields4 := eventMixin[4].Fields()
	_ = eventMixinFields4
	eventMixinFields5 := eventMixin[5].Fields()
	_ = eventMixinFields5
	eventFields := schema.Event{}.Fields()
	_ = eventFields
	// eventDescChannelID is the schema descriptor for channel_id field.
	eventDescChannelID := eventMixinFields2[0].Descriptor()
	// event.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	event.ChannelIDValidator = eventDescChannelID.Validators[0].(func(string) error)
	// eventDescPayload is the schema descriptor for payload field.
	eventDescPayload := eventMixinFields3[0].Descriptor()
	// event.DefaultPayload holds the default value on creation for the payload field.
	event.DefaultPayload = eventDescPayload.Default.(map[string]interface{})
	// eventDescUserID is the schema descriptor for user_id field.
	eventDescUserID := eventMixinFields4[0].Descriptor()
	// event.UserIDValidator is a validator for the "user_id" field. It is called by the builders before save.
	event.UserIDValidator = eventDescUserID.Validators[0].(func(string) error)
	// eventDescCreatedAt is the schema descriptor for created_at field.
	eventDescCreatedAt := eventMixinFields5[0].Descriptor()
	// event.DefaultCreatedAt holds the default value on creation for the created_at field.
	event.DefaultCreatedAt = eventDescCreatedAt.Default.(func() int64)
	// eventDescID is the schema descriptor for id field.
	eventDescID := eventMixinFields0[0].Descriptor()
	// event.DefaultID holds the default value on creation for the id field.
	event.DefaultID = eventDescID.Default.(func() string)
	// event.IDValidator is a validator for the "id" field. It is called by the builders before save.
	event.IDValidator = eventDescID.Validators[0].(func(string) error)
	notificationMixin := schema.Notification{}.Mixin()
	notificationMixinFields0 := notificationMixin[0].Fields()
	_ = notificationMixinFields0
	notificationMixinFields4 := notificationMixin[4].Fields()
	_ = notificationMixinFields4
	notificationMixinFields5 := notificationMixin[5].Fields()
	_ = notificationMixinFields5
	notificationMixinFields6 := notificationMixin[6].Fields()
	_ = notificationMixinFields6
	notificationMixinFields7 := notificationMixin[7].Fields()
	_ = notificationMixinFields7
	notificationMixinFields8 := notificationMixin[8].Fields()
	_ = notificationMixinFields8
	notificationFields := schema.Notification{}.Fields()
	_ = notificationFields
	// notificationDescUserID is the schema descriptor for user_id field.
	notificationDescUserID := notificationMixinFields4[0].Descriptor()
	// notification.UserIDValidator is a validator for the "user_id" field. It is called by the builders before save.
	notification.UserIDValidator = notificationDescUserID.Validators[0].(func(string) error)
	// notificationDescStatus is the schema descriptor for status field.
	notificationDescStatus := notificationMixinFields5[0].Descriptor()
	// notification.DefaultStatus holds the default value on creation for the status field.
	notification.DefaultStatus = notificationDescStatus.Default.(int)
	// notificationDescLinks is the schema descriptor for links field.
	notificationDescLinks := notificationMixinFields6[0].Descriptor()
	// notification.DefaultLinks holds the default value on creation for the links field.
	notification.DefaultLinks = notificationDescLinks.Default.([]map[string]interface{})
	// notificationDescChannelID is the schema descriptor for channel_id field.
	notificationDescChannelID := notificationMixinFields7[0].Descriptor()
	// notification.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	notification.ChannelIDValidator = notificationDescChannelID.Validators[0].(func(string) error)
	// notificationDescCreatedAt is the schema descriptor for created_at field.
	notificationDescCreatedAt := notificationMixinFields8[0].Descriptor()
	// notification.DefaultCreatedAt holds the default value on creation for the created_at field.
	notification.DefaultCreatedAt = notificationDescCreatedAt.Default.(func() int64)
	// notificationDescUpdatedAt is the schema descriptor for updated_at field.
	notificationDescUpdatedAt := notificationMixinFields8[1].Descriptor()
	// notification.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	notification.DefaultUpdatedAt = notificationDescUpdatedAt.Default.(func() int64)
	// notification.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	notification.UpdateDefaultUpdatedAt = notificationDescUpdatedAt.UpdateDefault.(func() int64)
	// notificationDescID is the schema descriptor for id field.
	notificationDescID := notificationMixinFields0[0].Descriptor()
	// notification.DefaultID holds the default value on creation for the id field.
	notification.DefaultID = notificationDescID.Default.(func() string)
	// notification.IDValidator is a validator for the "id" field. It is called by the builders before save.
	notification.IDValidator = notificationDescID.Validators[0].(func(string) error)
	subscriptionMixin := schema.Subscription{}.Mixin()
	subscriptionMixinFields0 := subscriptionMixin[0].Fields()
	_ = subscriptionMixinFields0
	subscriptionMixinFields1 := subscriptionMixin[1].Fields()
	_ = subscriptionMixinFields1
	subscriptionMixinFields2 := subscriptionMixin[2].Fields()
	_ = subscriptionMixinFields2
	subscriptionMixinFields3 := subscriptionMixin[3].Fields()
	_ = subscriptionMixinFields3
	subscriptionMixinFields4 := subscriptionMixin[4].Fields()
	_ = subscriptionMixinFields4
	subscriptionFields := schema.Subscription{}.Fields()
	_ = subscriptionFields
	// subscriptionDescUserID is the schema descriptor for user_id field.
	subscriptionDescUserID := subscriptionMixinFields1[0].Descriptor()
	// subscription.UserIDValidator is a validator for the "user_id" field. It is called by the builders before save.
	subscription.UserIDValidator = subscriptionDescUserID.Validators[0].(func(string) error)
	// subscriptionDescChannelID is the schema descriptor for channel_id field.
	subscriptionDescChannelID := subscriptionMixinFields2[0].Descriptor()
	// subscription.ChannelIDValidator is a validator for the "channel_id" field. It is called by the builders before save.
	subscription.ChannelIDValidator = subscriptionDescChannelID.Validators[0].(func(string) error)
	// subscriptionDescStatus is the schema descriptor for status field.
	subscriptionDescStatus := subscriptionMixinFields3[0].Descriptor()
	// subscription.DefaultStatus holds the default value on creation for the status field.
	subscription.DefaultStatus = subscriptionDescStatus.Default.(int)
	// subscriptionDescCreatedAt is the schema descriptor for created_at field.
	subscriptionDescCreatedAt := subscriptionMixinFields4[0].Descriptor()
	// subscription.DefaultCreatedAt holds the default value on creation for the created_at field.
	subscription.DefaultCreatedAt = subscriptionDescCreatedAt.Default.(func() int64)
	// subscriptionDescUpdatedAt is the schema descriptor for updated_at field.
	subscriptionDescUpdatedAt := subscriptionMixinFields4[1].Descriptor()
	// subscription.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	subscription.DefaultUpdatedAt = subscriptionDescUpdatedAt.Default.(func() int64)
	// subscription.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	subscription.UpdateDefaultUpdatedAt = subscriptionDescUpdatedAt.UpdateDefault.(func() int64)
	// subscriptionDescID is the schema descriptor for id field.
	subscriptionDescID := subscriptionMixinFields0[0].Descriptor()
	// subscription.DefaultID holds the default value on creation for the id field.
	subscription.DefaultID = subscriptionDescID.Default.(func() string)
	// subscription.IDValidator is a validator for the "id" field. It is called by the builders before save.
	subscription.IDValidator = subscriptionDescID.Validators[0].(func(string) error)
}
