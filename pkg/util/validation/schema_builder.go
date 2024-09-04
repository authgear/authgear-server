package validation

import (
	"encoding/json"
)

// SchemaBuilder is just map[string]interface{} with
// some convenient methods to build schema interactively.
// The zero value of a SchemaBuilder is an empty schema (by definition).
// This is NOT meant to be complete.
// It just supports the internal use cases we have.
type SchemaBuilder map[string]interface{}

type Type string

const (
	TypeNull    Type = "null"
	TypeBoolean Type = "boolean"
	TypeNumber  Type = "number"
	TypeInteger Type = "integer"
	TypeString  Type = "string"
	TypeObject  Type = "object"
	TypeArray   Type = "array"
)

func (b SchemaBuilder) Type(t Type) SchemaBuilder {
	b["type"] = t
	return b
}

func (b SchemaBuilder) Types(ts ...Type) SchemaBuilder {
	b["type"] = ts
	return b
}

func (b SchemaBuilder) AddTypeNull() SchemaBuilder {
	bType, ok := b["type"]
	if !ok {
		b.Type(TypeNull)
		return b
	}

	originals, ok := bType.([]Type)
	if ok {
		newTypes := append(originals, TypeNull)
		b.Types(newTypes...)
		return b
	}

	original, ok := bType.(Type)
	if ok {
		newTypes := []Type{original, TypeNull}
		b.Types(newTypes...)
		return b
	}

	panic("unexpected: schema builder has invalid type")
}

func (b SchemaBuilder) Properties() SchemaBuilder {
	bb, ok := b["properties"].(SchemaBuilder)
	if !ok {
		bb = SchemaBuilder{}
		b["properties"] = bb
	}
	return bb
}

func (b SchemaBuilder) Items(builder SchemaBuilder) SchemaBuilder {
	b["items"] = builder
	return b
}

func (b SchemaBuilder) Contains(builder SchemaBuilder) SchemaBuilder {
	b["contains"] = builder
	return b
}

func (b SchemaBuilder) Required(keys ...string) SchemaBuilder {
	b["required"] = keys
	return b
}

func (b SchemaBuilder) AddRequired(keys ...string) SchemaBuilder {
	originals, ok := b["required"].([]string)
	if ok {
		newRequired := append(originals, keys...)
		b["required"] = newRequired
	} else {
		b["required"] = keys
	}
	return b
}

func (b SchemaBuilder) Enum(values ...interface{}) SchemaBuilder {
	b["enum"] = values
	return b
}

func (b SchemaBuilder) Const(value interface{}) SchemaBuilder {
	b["const"] = value
	return b
}

func (b SchemaBuilder) Format(format string) SchemaBuilder {
	b["format"] = format
	return b
}

func (b SchemaBuilder) MinLength(minLength int) SchemaBuilder {
	b["minLength"] = minLength
	return b
}

func (b SchemaBuilder) MinimumInt64(minimum int64) SchemaBuilder {
	b["minimum"] = minimum
	return b
}

func (b SchemaBuilder) MaximumInt64(maximum int64) SchemaBuilder {
	b["maximum"] = maximum
	return b
}

func (b SchemaBuilder) MinimumFloat64(minimum float64) SchemaBuilder {
	b["minimum"] = minimum
	return b
}

func (b SchemaBuilder) MaximumFloat64(maximum float64) SchemaBuilder {
	b["maximum"] = maximum
	return b
}

func (b SchemaBuilder) AdditionalPropertiesFalse() SchemaBuilder {
	b["additionalProperties"] = false
	return b
}

func (b SchemaBuilder) Property(key string, builder SchemaBuilder) SchemaBuilder {
	b[key] = builder
	return b
}

func (b SchemaBuilder) OneOf(builders ...SchemaBuilder) SchemaBuilder {
	b["oneOf"] = builders
	return b
}

func (b SchemaBuilder) AllOf(builders ...SchemaBuilder) SchemaBuilder {
	b["allOf"] = builders
	return b
}

func (b SchemaBuilder) If(builder SchemaBuilder) SchemaBuilder {
	b["if"] = builder
	return b
}

func (b SchemaBuilder) Then(builder SchemaBuilder) SchemaBuilder {
	b["then"] = builder
	return b
}

func (b SchemaBuilder) Else(builder SchemaBuilder) SchemaBuilder {
	b["else"] = builder
	return b
}

func (b SchemaBuilder) ToSimpleSchema() *SimpleSchema {
	bytes, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return NewSimpleSchema(string(bytes))
}

// This function allow copying the schema builder to a reference
// Expected usage is to avoid mutating original schema builder
func (b SchemaBuilder) Clone() SchemaBuilder {
	newB := make(SchemaBuilder)
	for k, v := range b {
		vb, ok := v.(SchemaBuilder)
		if ok {
			newB[k] = vb.Clone()
		} else {
			newB[k] = v
		}
	}

	return newB
}
