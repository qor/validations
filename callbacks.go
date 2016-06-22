package validations

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"regexp"
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
					for _, err := range validatorErrors.(govalidator.Errors).Errors() {
						if strings.Index(err.Error(), "non zero value required") >= 0 {
							scope.DB().AddError(errors.New(fmt.Sprintf("%v can't be blank", err.(govalidator.Error).Name)))
						} else if strings.Index(err.Error(), "as length") >= 0 {
							reg, _ := regexp.Compile(`\(([0-9]+)\|([0-9]+)\)`)
							submatch := reg.FindSubmatch([]byte(err.Error()))
							scope.DB().AddError(errors.New(fmt.Sprintf("%v is the wrong length (should be %v~%v characters)", "Password", string(submatch[1]), string(submatch[2]))))
						} else if strings.Index(err.Error(), "as numeric") >= 0 {
							scope.DB().AddError(errors.New(fmt.Sprintf("%v is not a number", err.(govalidator.Error).Name)))
						} else if strings.Index(err.Error(), "as email") >= 0 {
							scope.DB().AddError(errors.New(fmt.Sprintf("%v is not a valid email address", err.(govalidator.Error).Name)))
						} else {
							scope.DB().AddError(errors.New(err.Error()))
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
