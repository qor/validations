package validations_test

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/test/utils"
	"github.com/qor/validations"
	"regexp"
	"testing"
)

var db *gorm.DB

type User struct {
	gorm.Model
	Name       string `valid:"required"`
	Password   string `valid:"length(6|10)"`
	CompanyID  int
	Company    Company
	CreditCard CreditCard
	Addresses  []Address
	Languages  []Language `gorm:"many2many:user_languages"`
}

func (user *User) Validate(db *gorm.DB) {
	if user.Name == "invalid" {
		db.AddError(validations.NewError(user, "Name", "invalid user name"))
	}
}

type Company struct {
	gorm.Model
	Name string
}

func (company *Company) Validate(db *gorm.DB) {
	if company.Name == "invalid" {
		db.AddError(errors.New("invalid company name"))
	}
}

type CreditCard struct {
	gorm.Model
	UserID int
	Number string
}

func (card *CreditCard) Validate(db *gorm.DB) {
	if !regexp.MustCompile("^(\\d){13,16}$").MatchString(card.Number) {
		db.AddError(validations.NewError(card, "Number", "invalid card number"))
	}
}

type Address struct {
	gorm.Model
	UserID  int
	Address string
}

func (address *Address) Validate(db *gorm.DB) {
	if address.Address == "invalid" {
		db.AddError(validations.NewError(address, "Address", "invalid address"))
	}
}

type Language struct {
	gorm.Model
	Code string
}

func (language *Language) Validate(db *gorm.DB) error {
	if language.Code == "invalid" {
		return validations.NewError(language, "Code", "invalid language")
	}
	return nil
}

func init() {
	db = utils.TestDB()
	validations.RegisterCallbacks(db)
	db.AutoMigrate(&User{}, &Company{}, &CreditCard{}, &Address{}, &Language{})
}

func TestGoValidation(t *testing.T) {
	user := User{Name: "", Password: "123123"}

	result := db.Save(&user)
	if result.Error == nil {
		t.Errorf("Should get error when save empty user")
	}

	if result.Error.Error() != "Name can't be blank" {
		t.Errorf("Error message should be equal `Name can't be blank`")
	}

	user = User{Name: "", Password: "123"}
	result = db.Save(&user)
	messages := []string{"Name can't be blank", "Password: 123 does not validate as length(6|10)"}
	for i, err := range result.Error.(gorm.Errors).GetErrors() {
		if messages[i] != err.Error() {
			t.Errorf(fmt.Sprintf("Error message should be equal `%v`", messages[i]))
		}
	}
}

func TestSaveInvalidUser(t *testing.T) {
	user := User{Name: "invalid"}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid user")
	}
}

func TestSaveInvalidCompany(t *testing.T) {
	user := User{
		Name:    "valid",
		Company: Company{Name: "invalid"},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid company")
	}
}

func TestSaveInvalidCreditCard(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "invalid"},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid credit card")
	}
}

func TestSaveInvalidAddresses(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "invalid"}},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid addresses")
	}
}

func TestSaveInvalidLanguage(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "valid"}},
		Languages:  []Language{{Code: "invalid"}},
	}

	if result := db.Save(&user); result.Error == nil {
		t.Errorf("Should get error when save invalid language")
	}
}

func TestSaveAllValidData(t *testing.T) {
	user := User{
		Name:       "valid",
		Company:    Company{Name: "valid"},
		CreditCard: CreditCard{Number: "4111111111111111"},
		Addresses:  []Address{{Address: "valid1"}, {Address: "valid2"}},
		Languages:  []Language{{Code: "valid1"}, {Code: "valid2"}},
	}

	if result := db.Save(&user); result.Error != nil {
		t.Errorf("Should get no error when save valid data, but got: %v", result.Error)
	}
}
