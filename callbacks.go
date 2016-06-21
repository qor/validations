package validations

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"strings"
)

var skipValidations = "validations:skip_validations"

func validate(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
			if !scope.HasError() {
				scope.CallMethod("Validate")
				_, validatorErrors := govalidator.ValidateStruct(scope.IndirectValue().Interface())
				if validatorErrors != nil {
					var errorsArr govalidator.Errors
					if e, ok := validatorErrors.(govalidator.Errors); ok {
						errorsArr = e
					} else {
						errorsArr = append(errorsArr, e)
					}
					for _, err := range errorsArr.Errors() {
						validatorError := err.(govalidator.Error)
						if strings.Index(validatorError.Error(), "non zero value required") >= 0 {
							scope.DB().AddError(errors.New(fmt.Sprintf("%v can't be blank", err.(govalidator.Error).Name)))
						} else {
							scope.DB().AddError(errors.New(validatorError.Error()))
						}
					}
				}
			}
		}
	}
}

// RegisterCallbacks register callback into GORM DB
func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	callback.Create().Before("gorm:before_create").Register("validations:validate", validate)
	callback.Update().Before("gorm:before_update").Register("validations:validate", validate)
}
