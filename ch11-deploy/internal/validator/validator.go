package validator

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	ja_locale "github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

// ErrEngineNotValidator は Gin の binding エンジンが
// *validator.Validate ではなかった場合に返すセンチネル
var ErrEngineNotValidator = errors.New(
	"binding.Validator.Engine() is not *validator.Validate")

// Setup は Gin 内蔵の validator を取り出して本書の設定を入れる
func Setup() (*validator.Validate, error) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil, ErrEngineNotValidator
	}
	registerJSONTagName(v)
	if err := registerPostalCode(v); err != nil {
		return nil, err
	}
	return v, nil
}

// registerJSONTagName は validator のエラーメッセージのフィールド名を
// Go の struct フィールド名 (PascalCase) から JSON タグ名 (snake_case) に置き換える
func registerJSONTagName(v *validator.Validate) {
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

var postalCodeRegex = regexp.MustCompile(`^\d{3}-\d{4}$`)

// registerPostalCode は postal_code_jp タグを追加する
// 例: Zip string `json:"zip" binding:"required,postal_code_jp"`
func registerPostalCode(v *validator.Validate) error {
	return v.RegisterValidation(
		"postal_code_jp",
		func(fl validator.FieldLevel) bool {
			return postalCodeRegex.MatchString(fl.Field().String())
		},
	)
}

// NewJapaneseTranslator は日本語翻訳つきの Translator を返す
func NewJapaneseTranslator(
	v *validator.Validate,
) (ut.Translator, error) {
	uni := ut.New(ja_locale.New(), ja_locale.New())
	trans, _ := uni.GetTranslator("ja")
	if err := ja_translations.RegisterDefaultTranslations(v, trans); err != nil {
		return nil, err
	}
	return trans, nil
}

// RegisterPostalCodeJA はカスタムタグ postal_code_jp の日本語訳を登録する
func RegisterPostalCodeJA(
	v *validator.Validate, trans ut.Translator,
) error {
	return v.RegisterTranslation("postal_code_jp", trans,
		func(ut ut.Translator) error {
			return ut.Add("postal_code_jp",
				"{0}は 000-0000 の形式で入力してください", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("postal_code_jp", fe.Field())
			return t
		},
	)
}
