package validator

import (
	"reflect"

	"github.com/gin-gonic/gin/binding"
	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	validator  *validator.Validate
	Translator ut.Translator
}

func (v *Validator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		if value.Elem().Kind() != reflect.Struct {
			return v.ValidateStruct(value.Elem().Interface())
		}
		return v.validator.Struct(obj)
	case reflect.Struct:
		return v.validator.Struct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(binding.SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

func (v *Validator) Engine() any {
	return v.validator
}

func NewValidator() (*Validator, error) {
	v := validator.New()
	v.SetTagName("binding")

	enLocale := en.New()
	uni := ut.New(enLocale)
	trans, _ := uni.GetTranslator("en")

	if err := entranslations.RegisterDefaultTranslations(v, trans); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("gtZero", validatePrice); err != nil {
		return nil, err
	}
	if err := v.RegisterTranslation(
		"gtZero",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("gtZero", "{0} must be greater than zero", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtZero", fe.Field())
			return t
		},
	); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("orderStatus", validateOrderSessionStatus); err != nil {
		return nil, err
	}
	if err := v.RegisterTranslation(
		"orderStatus",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("orderStatus", "{0} is not a valid order status", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("orderStatus", fe.Field())
			return t
		},
	); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("orderedProductStatus", validateOrderedProductStatus); err != nil {
		return nil, err
	}
	if err := v.RegisterTranslation(
		"orderedProductStatus",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("orderedProductStatus", "{0} is not a valid ordered product status", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("orderedProductStatus", fe.Field())
			return t
		},
	); err != nil {
		return nil, err
	}

	return &Validator{
		validator:  v,
		Translator: trans,
	}, nil
}
