package schema

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/byebyebymyai/oauth2-api/ent/schema/uuidgql"
)

// Oauth2Client holds the schema definition for the Oauth2Client entity.
type Oauth2Client struct {
	ent.Schema
}

// Fields of the Oauth2Client.
func (Oauth2Client) Fields() []ent.Field {
	return []ent.Field{
		field.String("secret").NotEmpty().Annotations(entproto.Field(2)),
		field.String("domain").NotEmpty().Annotations(entproto.Field(3)),
	}
}

// Edges of the Oauth2Client.
func (Oauth2Client) Edges() []ent.Edge {
	return nil
}

// Mixin returns User mixed-in schema.
func (Oauth2Client) Mixin() []ent.Mixin {
	return []ent.Mixin{
		uuidgql.MixinWithID(),
	}
}

// Annotations returns Oauth2Client annotations.
func (Oauth2Client) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
