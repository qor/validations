package validations

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

const skipValidations = "validations:skip_validations"

func validate(db *gorm.DB) {
	if _, ok := db.Get("gorm:update_column"); ok {
		return
	}
	result, ok := db.Get(skipValidations)
	if ok && result.(bool) {
		return
	}
	if db.Error != nil {
		return
	}
	if method, ok := db.Statement.Schema.ModelType.MethodByName("Validate"); ok {
		method.Func.Call([]reflect.Value{})
	}
	if db.Statement != nil {
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				resource := db.Statement.ReflectValue.Index(i).Interface()
				_, validatorErrors := govalidator.ValidateStruct(resource)
				if validatorErrors == nil {
					return
				}
				if errors, ok := validatorErrors.(govalidator.Errors); ok {
					for _, err := range flatValidatorErrors(errors) {
						_ = db.AddError(formattedError(err, resource))
					}
				} else {
					_ = db.AddError(validatorErrors)
				}
			}
		case reflect.Struct:
			resource := db.Statement.ReflectValue.Interface()
			_, validatorErrors := govalidator.ValidateStruct(resource)
			if validatorErrors == nil {
				return
			}
			if errors, ok := validatorErrors.(govalidator.Errors); ok {
				for _, err := range flatValidatorErrors(errors) {
					_ = db.AddError(formattedError(err, resource))
				}
			} else {
				_ = db.AddError(validatorErrors)
			}
		}
	}
}

func flatValidatorErrors(validatorErrors govalidator.Errors) []govalidator.Error {
	resultErrors := make([]govalidator.Error, 0)
	for _, validatorError := range validatorErrors.Errors() {
		if errors, ok := validatorError.(govalidator.Errors); ok {
			for _, e := range errors {
				resultErrors = append(resultErrors, e.(govalidator.Error))
			}
		}
		if e, ok := validatorError.(govalidator.Error); ok {
			resultErrors = append(resultErrors, e)
		}
	}
	return resultErrors
}

func formattedError(err govalidator.Error, resource interface{}) error {
	message := err.Error()
	attrName := err.Name
	if strings.Index(message, "non zero value required") >= 0 {
		message = fmt.Sprintf("%v can't be blank", attrName)
	} else if strings.Index(message, "as length") >= 0 {
		reg, _ := regexp.Compile(`\(([0-9]+)\|([0-9]+)\)`)
		submatch := reg.FindSubmatch([]byte(err.Error()))
		message = fmt.Sprintf("%v is the wrong length (should be %v~%v characters)", attrName, string(submatch[1]), string(submatch[2]))
	} else if strings.Index(message, "as numeric") >= 0 {
		message = fmt.Sprintf("%v is not a number", attrName)
	} else if strings.Index(message, "as email") >= 0 {
		message = fmt.Sprintf("%v is not a valid email address", attrName)
	}
	return NewError(resource, attrName, message)

}

// RegisterCallbacks register callbacks into GORM DB
func RegisterCallbacks(db *gorm.DB) error {
	callback := db.Callback()
	if callback.Create().Get("validations:validate") == nil {
		err := callback.Create().Before("gorm:before_create").Register("validations:validate", validate)
		if err != nil {
			return err
		}
	}
	if callback.Update().Get("validations:validate") == nil {
		err := callback.Update().Before("gorm:before_update").Register("validations:validate", validate)
		if err != nil {
			return err
		}
	}
	return nil
}
