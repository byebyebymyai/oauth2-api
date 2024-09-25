// Copyright 2019-present Facebook
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uuidgql

import (
	"entgo.io/contrib/entgql"
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// MixinWithID creates a Mixin that encodes the provided prefix.
func MixinWithID() *Mixin {
	return &Mixin{}
}

// Mixin defines an ent Mixin that captures the PULID prefix for a type.
type Mixin struct {
	mixin.Schema
}

// Fields provides the id field.
func (m Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Annotations(entproto.Field(1)),
	}
}

// Annotation captures the id prefix for a type.
type Annotation struct {
}

// Name implements the ent Annotation interface.
func (a Annotation) Name() string {
	return "UUID"
}

// Annotations returns the annotations for a Mixin instance.
func (m Mixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entproto.Message(),
		entproto.Service(),
	}
}
