package validations

import (
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
						message := err.Error()
						attrName := err.(govalidator.Error).Name
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
						scope.DB().AddError(NewError(scope.IndirectValue().Interface(), attrName, message))
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
