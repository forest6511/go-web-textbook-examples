package apperror

// FieldIssue はフィールド単位のバリデーション詳細
type FieldIssue struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

// AppError は HTTP 境界で扱うエラー表現
type AppError struct {
	Code     string       // 安定コード: "VALIDATION_FAILED" 等
	Message  string       // 利用者向けメッセージ
	HTTPCode int          // HTTP ステータス
	Cause    error        // 下位エラー (%w 相当)
	Details  []FieldIssue // バリデーション詳細
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + " (" + e.Cause.Error() + ")"
	}
	return e.Code + ": " + e.Message
}

// Unwrap で errors.Is / errors.As の連鎖を通す
func (e *AppError) Unwrap() error { return e.Cause }
