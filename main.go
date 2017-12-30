package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"math"
)
const(
	DefaultMaxValue = math.MaxInt8
	DefaultMinValue = math.MinInt8
)
type Validator interface {
	Validate(value interface{}) (bool, error)
}
type Option struct{
	min int
	max int
	pattern string
}

type OptionFunc func(*Option)

func WithMin(min int)OptionFunc{
	return func(option *Option) {
		option.min = min
	}
}

func WithMax(max int)OptionFunc{
	return func(option *Option){
		option.max = max
	}
}

func WithPattern(pattern string)OptionFunc{
	return func(option *Option){
		option.pattern = pattern
	}
}

var ValidatorFactory = make(map[string]func(...OptionFunc)Validator)

func init() {
	ValidatorFactory["default"] = newDefaultValidator
	ValidatorFactory["int"] = newNumberValidator
	ValidatorFactory["string"] = newStringValidator
	ValidatorFactory["email"] = newEmailValidator
}

type DefaultValidator struct{}

func newDefaultValidator(...OptionFunc)Validator{
	return &DefaultValidator{}
}

func (d *DefaultValidator) Validate(val interface{}) (bool, error) {
	return true, nil
}

type StringValidator struct {
	min, max int
}

func newStringValidator(opts ...OptionFunc)Validator{
	option := &Option{}
	for _,opt := range opts{
		opt(option)
	}
	strValidator := &StringValidator{}
	strValidator.min = option.min
	strValidator.max = option.max
	return strValidator
}

func (s *StringValidator) Validate(val interface{}) (bool, error) {
	str, ok := val.(string)
	if !ok {
		return false, fmt.Errorf("not of string type")
	}

	if strLen := len(str); strLen < s.min || strLen > s.max {
		return false, fmt.Errorf("string length %d, allowed range [ %d,%d ]", strLen, s.min, s.max)
	}
	return true, nil
}

type NumberValidator struct {
	min, max int
}

func newNumberValidator(opts ...OptionFunc)Validator{
	option := &Option{}
	for _,opt := range opts{
		opt(option)
	}
	numValidator := &NumberValidator{}
	numValidator.max = option.max
	numValidator.min = option.min
	return numValidator
}

func (n *NumberValidator) Validate(val interface{}) (bool, error) {
	num, ok := val.(int)
	if !ok {
		return false, errors.New("not of int type")
	}

	if num < n.min || num > n.max {
		return false, fmt.Errorf("interger %d, allowed range [ %d,%d]", num, n.min, n.max)
	}
	return true, nil
}

type EmailValidator struct {
	pattern string
}

func newEmailValidator(opts ...OptionFunc)Validator{
	option := &Option{}
	for _,opt := range opts{
		opt(option)
	}
	emailValidator := &EmailValidator{}
	emailValidator.pattern = option.pattern
	return emailValidator
}

func (e *EmailValidator) Validate(email interface{}) (bool, error) {
	regExp := regexp.MustCompile(e.pattern)
	if !regExp.MatchString(email.(string)) {
		return false, errors.New("not a valid email address")
	}
	return true, nil
}

var tagName = `validate`
var pattern = `\A[\w+\-.]+@[a-z\d\-]+(\.[a-z]+)*\.[a-z]+\z`

type User struct {
	Name      string `validate:"string"`
	Email     string `validate:"email"`
	Age       int    `validate:"int,min=18,max=30"`
	ContactNo string `validate:"string,min=10,max=13"`
}

func GetValidatorFromTag(tag string) (Validator, error) {
	args := strings.Split(tag, ",")
	var min, max int
	if len(args) == 0 {
		return nil, fmt.Errorf("validator type not present")
	}
	validator, ok := ValidatorFactory[args[0]]
	if !ok {
		return nil, fmt.Errorf("validator for %s not present", args[0])
	}

	if len(args) == 1 {
		min = int(DefaultMinValue)
		max = int(DefaultMaxValue)
	} else {
		fmt.Sscanf(strings.Join(args[1:], ","), "min=%d,max=%d", &min, &max)
	}

	return validator(WithMin(min),WithMax(max),WithPattern(pattern)),nil
}

func ValidateUser(user interface{}) []error {
	errs := make([]error, 0)
	val := reflect.ValueOf(user)
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag.Get(tagName)
		if tag == "" || tag == "_"{
			continue
		}
		validator,err := GetValidatorFromTag(tag)
		if err != nil {
			errs = append(errs,err)
			continue
		}
		valid,err := validator.Validate(val.Field(i).Interface())
		if err != nil && !valid{
			errs = append(errs,err)
		}
	}
	return errs
}

func main() {
	user := User{
		Name:      "oshank",
		Email:     "oshankfriends@gmail.com",
		Age:       85,
		ContactNo: "7065349354",
	}
	fmt.Println("Errors:")
	for i,err := range ValidateUser(user){
		fmt.Printf("%d. %s\n",i+1,err.Error())
	}
}
