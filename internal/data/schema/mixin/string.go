package mixin

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Email adds email field.
type Email struct{ mixin.Schema }

// Fields of the email mixin.
func (Email) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Comment("email").Optional(), // email
	}
}

// email mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Email)(nil)

// Username adds username field.
type Username struct{ mixin.Schema }

// Fields of the username mixin.
func (Username) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Comment("username").Optional(), // username
	}
}

// username mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Username)(nil)

// UsernameUnique adds username field.
type UsernameUnique struct{ mixin.Schema }

// Fields of the username mixin.
func (UsernameUnique) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Comment("username").Unique().NotEmpty().Optional().MaxLen(50).Match(regexp.MustCompile("^[a-zA-Z0-9]{4,16}$")), // username
	}
}

// Indexes of the Username.
func (UsernameUnique) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
	}
}

// username mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UsernameUnique)(nil)

// Password adds password field.
type Password struct{ mixin.Schema }

// Fields of the password mixin.
func (Password) Fields() []ent.Field {
	return []ent.Field{
		field.String("password").Comment("password").Sensitive().Optional(), // password
	}
}

// password mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Password)(nil)

// Secret adds secret field.
type Secret struct{ mixin.Schema }

// Fields of the secret mixin.
func (Secret) Fields() []ent.Field {
	return []ent.Field{
		field.String("secret").Comment("secret key").Optional(), // secret key
	}
}

// secret mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Secret)(nil)

// Phone adds phone field.
type Phone struct{ mixin.Schema }

// Fields of the phone mixin.
func (Phone) Fields() []ent.Field {
	return []ent.Field{
		field.String("phone").Comment("phone").Optional(), // phone
	}
}

// phone mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Phone)(nil)

// BankName adds expired at time field.
type BankName struct{ mixin.Schema }

// Fields of the bank name mixin.
func (BankName) Fields() []ent.Field {
	return []ent.Field{
		field.String("bank_name").Comment("bank name").Optional(), // Bank Name
	}
}

// bank name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*BankName)(nil)

// CardNo adds card no field.
type CardNo struct{ mixin.Schema }

// Fields of the card no mixin.
func (CardNo) Fields() []ent.Field {
	return []ent.Field{
		field.String("card_no").Comment("card no").Optional(), // Card No
	}
}

// card no mixin must implement `Mixin` interface.
var _ ent.Mixin = (*CardNo)(nil)

// CCV adds ccv field.
type CCV struct{ mixin.Schema }

// Fields of the ccv mixin.
func (CCV) Fields() []ent.Field {
	return []ent.Field{
		field.String("ccv").Comment("ccv").Optional(), // ccv
	}
}

// ccv mixin must implement `Mixin` interface.
var _ ent.Mixin = (*CCV)(nil)

// Province adds province field.
type Province struct{ mixin.Schema }

// Fields of the province mixin.
func (Province) Fields() []ent.Field {
	return []ent.Field{
		field.String("province").Comment("province").Optional(), // province
	}
}

// province mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Province)(nil)

// ZipCode adds expired at time field.
type ZipCode struct{ mixin.Schema }

// Fields of the zip code mixin.
func (ZipCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("zip_code").Comment("zip code").Optional(), // Zip code
	}
}

// City adds city field.
type City struct{ mixin.Schema }

// Fields of the city mixin.
func (City) Fields() []ent.Field {
	return []ent.Field{
		field.String("city").Comment("city").Optional(), // city
	}
}

// city mixin must implement `Mixin` interface.
var _ ent.Mixin = (*City)(nil)

// District adds district field.
type District struct{ mixin.Schema }

// Fields of the district mixin.
func (District) Fields() []ent.Field {
	return []ent.Field{
		field.String("district").Comment("district").Optional(), // district
	}
}

// district mixin must implement `Mixin` interface.
var _ ent.Mixin = (*District)(nil)

// Address adds address field.
type Address struct{ mixin.Schema }

// Fields of the address mixin.
func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.String("address").Comment("address").Optional(), // address
	}
}

// address mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Address)(nil)

// zip code mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ZipCode)(nil)

// FirstName adds first name field.
type FirstName struct{ mixin.Schema }

// Fields of the first name mixin.
func (FirstName) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name").Comment("first name").Optional(), // first name
	}
}

// first name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*FirstName)(nil)

// LastName adds last name field.
type LastName struct{ mixin.Schema }

// Fields of the last name mixin.
func (LastName) Fields() []ent.Field {
	return []ent.Field{
		field.String("last_name").Comment("last name").Optional(), // last name
	}
}

// last name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*LastName)(nil)

// DisplayName adds display name field.
type DisplayName struct{ mixin.Schema }

// Fields of the display name mixin.
func (DisplayName) Fields() []ent.Field {
	return []ent.Field{
		field.String("display_name").Comment("display name").Optional(), // display name
	}
}

// display name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*DisplayName)(nil)

// Language adds language field.
type Language struct{ mixin.Schema }

// Fields of the language mixin.
func (Language) Fields() []ent.Field {
	return []ent.Field{
		field.String("language").Comment("language").Optional(), // language
	}
}

// language mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Language)(nil)

// About adds about field.
type About struct{ mixin.Schema }

// Fields of the about mixin.
func (About) Fields() []ent.Field {
	return []ent.Field{
		field.String("about").Comment("about").Optional(), // about
	}
}

// about mixin must implement `Mixin` interface.
var _ ent.Mixin = (*About)(nil)

// Name adds name field.
type Name struct{ mixin.Schema }

// Fields of the name mixin.
func (Name) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("name").Optional(), // name
	}
}

// name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Name)(nil)

// Label adds label field.
type Label struct{ mixin.Schema }

// Fields of the label mixin.
func (Label) Fields() []ent.Field {
	return []ent.Field{
		field.String("label").Comment("label").Optional(), // label
	}
}

// label mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Label)(nil)

// Code adds code field.
type Code struct{ mixin.Schema }

// Fields of the code mixin.
func (Code) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").Comment("code").Optional(), // code
	}
}

// code mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Code)(nil)

// Slug adds slug field.
type Slug struct{ ent.Schema }

// Fields of the slug mixin.
func (Slug) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Comment("slug / alias").Optional(), // alias name
	}
}

// slug mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Slug)(nil)

// SlugUnique adds slug field.
type SlugUnique struct{ ent.Schema }

// Fields of the slug mixin.
func (SlugUnique) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").Unique().Comment("slug / alias").Optional(), // alias name
	}
}

// Indexes of the SlugUnique.
func (SlugUnique) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
	}
}

// slug mixin must implement `Mixin` interface.
var _ ent.Mixin = (*SlugUnique)(nil)

// Cover adds cover field.
type Cover struct{ mixin.Schema }

// Fields of the cover mixin.
func (Cover) Fields() []ent.Field {
	return []ent.Field{
		field.String("cover").Comment("cover").Optional(), // cover
	}
}

// cover mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Cover)(nil)

// Thumbnail adds thumbnail field.
type Thumbnail struct{ mixin.Schema }

// Fields of the thumbnail mixin.
func (Thumbnail) Fields() []ent.Field {
	return []ent.Field{
		field.String("thumbnail").Comment("thumbnail").Optional(), // thumbnail
	}
}

// thumbnail mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Thumbnail)(nil)

// Path adds path field.
type Path struct{ mixin.Schema }

// Fields of the path mixin.
func (Path) Fields() []ent.Field {
	return []ent.Field{
		field.String("path").Comment("path").Optional(), // path
	}
}

// path mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Path)(nil)

// Target adds target field.
type Target struct{ mixin.Schema }

// Fields of the target mixin.
func (Target) Fields() []ent.Field {
	return []ent.Field{
		field.String("target").Comment("target").Optional(), // target
	}
}

// target mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Target)(nil)

// URL adds url field.
type URL struct{ mixin.Schema }

// Fields of the url mixin.
func (URL) Fields() []ent.Field {
	return []ent.Field{
		field.String("url").Comment("url, website / link...").Optional(), // url
	}
}

// url mixin must implement `Mixin` interface.
var _ ent.Mixin = (*URL)(nil)

// Icon adds icon field.
type Icon struct{ mixin.Schema }

// Fields of the icon mixin.
func (Icon) Fields() []ent.Field {
	return []ent.Field{
		field.String("icon").Comment("icon").Optional(), // icon
	}
}

// icon mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Icon)(nil)

// Perms adds perms field.
type Perms struct{ mixin.Schema }

// Fields of the perms mixin.
func (Perms) Fields() []ent.Field {
	return []ent.Field{
		field.String("perms").Comment("perms").Optional(), // perms
	}
}

// perms mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Perms)(nil)

// Color adds color field.
type Color struct{ mixin.Schema }

// Fields of the color mixin.
func (Color) Fields() []ent.Field {
	return []ent.Field{
		field.String("color").Comment("color").Optional(), // color
	}
}

// color mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Color)(nil)

// Content adds content field.
type Content struct{ mixin.Schema }

// Fields of the content mixin.
func (Content) Fields() []ent.Field {
	return []ent.Field{
		field.Text("content").Comment("content, big text").Optional(),
	}
}

// content mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Content)(nil)

// Keywords adds keywords field.
type Keywords struct{ mixin.Schema }

// Fields of the keywords mixin.
func (Keywords) Fields() []ent.Field {
	return []ent.Field{
		field.String("keywords").Comment("keywords").Optional(), // keywords
	}
}

// keywords mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Keywords)(nil)

// Description adds description field.
type Description struct{ mixin.Schema }

// Fields of the description mixin.
func (Description) Fields() []ent.Field {
	return []ent.Field{
		field.Text("description").Comment("description").Optional(), // description
	}
}

// description mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Description)(nil)

// Copyright adds copyright field.
type Copyright struct{ mixin.Schema }

// Fields of the copyright mixin.
func (Copyright) Fields() []ent.Field {
	return []ent.Field{
		field.String("copyright").Comment("copyright").Optional(), // copyright
	}
}

// copyright mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Copyright)(nil)

// Logo adds logo field.
type Logo struct{ mixin.Schema }

// Fields of the logo mixin.
func (Logo) Fields() []ent.Field {
	return []ent.Field{
		field.String("logo").Comment("logo").Optional(), // logo
	}
}

// logo mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Logo)(nil)

// LogoAlt adds logo alt field.
type LogoAlt struct{ mixin.Schema }

// Fields of the logo alt mixin.
func (LogoAlt) Fields() []ent.Field {
	return []ent.Field{
		field.String("logo_alt").Comment("logo alt").Optional(), // logo alt
	}
}

// logo alt mixin must implement `Mixin` interface.
var _ ent.Mixin = (*LogoAlt)(nil)

// Type adds type field.
type Type struct{ mixin.Schema }

// Fields of the type mixin.
func (Type) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").Comment("type").Optional(), // type
	}
}

// type mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Type)(nil)

// Value adds value field.
type Value struct{ mixin.Schema }

// Fields of the value mixin.
func (Value) Fields() []ent.Field {
	return []ent.Field{
		field.String("value").Comment("value").Optional(), // value
	}
}

// value mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Value)(nil)

// Action adds action field.
type Action struct{ mixin.Schema }

// Fields of the action mixin.
func (Action) Fields() []ent.Field {
	return []ent.Field{
		field.String("action").Comment("action").Optional(), // action
	}
}

// action mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Action)(nil)

// Subject adds subject field.
type Subject struct{ mixin.Schema }

// Fields of the subject mixin.
func (Subject) Fields() []ent.Field {
	return []ent.Field{
		field.String("subject").Comment("subject").Optional(), // subject
	}
}

// subject mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Subject)(nil)

// Provider adds provider field.
type Provider struct{ ent.Schema }

// Fields of the provider mixin.
func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").Comment("provider").Optional(), // provider
	}
}

// provider mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Provider)(nil)

// AccessToken adds access token field.
type AccessToken struct{ ent.Schema }

// Fields of the access token mixin.
func (AccessToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("access_token").Comment("access token").NotEmpty(), // access token
	}
}

// access token mixin must implement `Mixin` interface.
var _ ent.Mixin = (*AccessToken)(nil)

// RefreshToken adds refresh token field.
type RefreshToken struct{ ent.Schema }

// Fields of the refresh token mixin.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("refresh_token").Comment("refresh token").NotEmpty(), // refresh token
	}
}

// refresh token mixin must implement `Mixin` interface.
var _ ent.Mixin = (*RefreshToken)(nil)

// SessionID adds session id field.
type SessionID struct{ mixin.Schema }

// Fields of the session id mixin.
func (SessionID) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").Comment("session id").Optional(), // session id
	}
}

// session id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*SessionID)(nil)

// ShortBio adds short bio field.
type ShortBio struct{ mixin.Schema }

// Fields of the short bio mixin.
func (ShortBio) Fields() []ent.Field {
	return []ent.Field{
		field.String("short_bio").Comment("short bio").Optional(), // short bio
	}
}

// short bio mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ShortBio)(nil)

// Bio adds bio field.
type Bio struct{ mixin.Schema }

// Fields of the bio mixin.
func (Bio) Fields() []ent.Field {
	return []ent.Field{
		field.Text("bio").Comment("bio").Optional(), // bio
	}
}

// bio mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Bio)(nil)

// Hash adds hash field.
type Hash struct{ mixin.Schema }

// Fields of the hash mixin.
func (Hash) Fields() []ent.Field {
	return []ent.Field{
		field.String("hash").Comment("hash").Optional(), // hash
	}
}

// hash mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Hash)(nil)

// Title adds title field.
type Title struct{ mixin.Schema }

// Fields of the title mixin.
func (Title) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").Comment("title").Optional(), // title
	}
}

// title mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Title)(nil)

// Caption adds caption field.
type Caption struct{ mixin.Schema }

// Fields of the caption mixin.
func (Caption) Fields() []ent.Field {
	return []ent.Field{
		field.String("caption").Comment("caption").Optional(), // caption
	}
}

// caption mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Caption)(nil)

// MediaType adds resource type field.
type MediaType struct{ mixin.Schema }

// Fields of the resource type mixin.
func (MediaType) Fields() []ent.Field {
	return []ent.Field{
		field.String("mime").Comment("resource type").Optional(), // resource type
	}
}

// resource type mixin must implement `Mixin` interface.
var _ ent.Mixin = (*MediaType)(nil)

// ExtensionName adds extension name field.
type ExtensionName struct{ mixin.Schema }

// Fields of the extension name mixin.
func (ExtensionName) Fields() []ent.Field {
	return []ent.Field{
		field.String("ext").Comment("extension name").Optional(), // extension name
	}
}

// extension name mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ExtensionName)(nil)
