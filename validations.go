package validations

import (
	"fmt"
	"gorm.io/gorm"
	"reflect"
)

// NewError generate a new error for a model's field
func NewError(resource interface{}, column, err string) error {
	return &Error{Resource: resource, Column: column, Message: err}
}

// Error is a validation error struct that hold model, column and error message
type Error struct {
	Resource interface{}
	Column   string
	Message  string
}

// Label is a label including model type, primary key and column name
func (err Error) Label() string {
	db := gorm.DB{
		Statement: &gorm.Statement{
			ReflectValue: reflect.ValueOf(err.Resource),
		},
	}
	if len(db.Statement.Schema.PrimaryFields) > 1 {
		v := make([]interface{}, 0)
		for _, f := range db.Statement.Schema.PrimaryFields {
			v = append(v, db.Statement.ReflectValue.FieldByName(f.Name))
		}
		return fmt.Sprintf("%v_%v_%v", db.Statement.Schema.ModelType.Name(), v, err.Column)
	} else {
		v := db.Statement.ReflectValue.FieldByName(db.Statement.Schema.PrimaryFields[0].Name)
		return fmt.Sprintf("%v_%v_%v", db.Statement.Schema.ModelType.Name(), v, err.Column)
	}
}

// Error show error message
func (err Error) Error() string {
	return fmt.Sprintf("%v", err.Message)
}
